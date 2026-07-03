// Semantic search system

// cSpell:ignore uuid, NewIDUUID

package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	sse_proto "github.com/AgustinSRG/PersonalMediaVault/semantic-search-engine/protocol/sse-proto-go"
	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Semantic search configuration
type SemanticSearchConfig struct {
	Enabled        bool
	SseBinPath     string
	ModelPath      string
	ImageSizeLimit uint64
}

const SSE_DEFAULT_SIZE_LIMIT_MB = 20

func LoadSemanticSearchConfig() *SemanticSearchConfig {
	if os.Getenv("SEMANTIC_SEARCH_ENABLED") != "YES" {
		return &SemanticSearchConfig{
			Enabled: false,
		}
	}

	sseBinPath := os.Getenv("SSE_BIN_PATH")

	if sseBinPath == "" {
		LogWarning("SSE_BIN_PATH is unset. Semantic search is disabled.")
		return &SemanticSearchConfig{
			Enabled: false,
		}
	}

	modelPath := os.Getenv("SSE_MODEL_PATH")

	if modelPath == "" {
		LogWarning("SSE_MODEL_PATH is unset. Semantic search is disabled.")
		return &SemanticSearchConfig{
			Enabled: false,
		}
	}

	imageSizeLimit := uint64(SSE_DEFAULT_SIZE_LIMIT_MB) * 1024 * 1024
	imageSizeLimitMbStr := os.Getenv("SSE_IMAGE_SIZE_LIMIT_MB")

	if imageSizeLimitMbStr != "" {
		v, err := strconv.ParseUint(imageSizeLimitMbStr, 10, 32)

		if err == nil {
			imageSizeLimit = v * 1024 * 1024
		} else {
			LogWarning("Error parsing SSE_IMAGE_SIZE_LIMIT_MB value: " + err.Error() + " | Using " + fmt.Sprint(SSE_DEFAULT_SIZE_LIMIT_MB) + " as fallback value.")
		}
	}

	return &SemanticSearchConfig{
		Enabled:        true,
		SseBinPath:     sseBinPath,
		ModelPath:      modelPath,
		ImageSizeLimit: imageSizeLimit,
	}
}

// Status for SemanticSearchSystem
type SemanticSearchSystemStatus struct {
	// Is the service available?
	available bool

	// GRPC client
	client sse_proto.SemanticSearchEngineServiceClient

	// Dimensions for the model
	dimensions uint

	// True if the initial scan was performed
	initialized bool
}

// Semantic search sub-system
type SemanticSearchSystem struct {
	// Configuration
	sseBinPath     string
	modelPath      string
	dbPath         string
	imageSizeLimit uint64
	apiKey         string

	// Status
	status   SemanticSearchSystemStatus
	statusMu *sync.Mutex

	// Pending state
	busy          map[uint64]*sync.WaitGroup
	pendingIndex  map[uint64]bool
	pendingDelete map[uint64]bool
	pendingMu     *sync.Mutex
}

func randomSeeApiKey() (string, error) {
	apiKey := make([]byte, 16)
	_, err := rand.Read(apiKey)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(apiKey), nil
}

// Creates instance of SemanticSearchSystem
func CreateSemanticSearchSystem(config *SemanticSearchConfig, vaultPath string) (*SemanticSearchSystem, error) {
	apiKey, err := randomSeeApiKey()

	if err != nil {
		return nil, err
	}

	return &SemanticSearchSystem{
		sseBinPath:     config.SseBinPath,
		modelPath:      config.ModelPath,
		dbPath:         path.Join(vaultPath, "semantic-search.db"),
		imageSizeLimit: config.ImageSizeLimit,
		apiKey:         apiKey,

		status: SemanticSearchSystemStatus{
			available:   false,
			client:      nil,
			dimensions:  0,
			initialized: false,
		},

		statusMu: &sync.Mutex{},

		busy:          make(map[uint64]*sync.WaitGroup),
		pendingIndex:  make(map[uint64]bool),
		pendingDelete: make(map[uint64]bool),
		pendingMu:     &sync.Mutex{},
	}, nil
}

// Gets the status of the sub-system
func (s *SemanticSearchSystem) GetStatus() SemanticSearchSystemStatus {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	return s.status
}

func (s *SemanticSearchSystem) GetClient() sse_proto.SemanticSearchEngineServiceClient {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	return s.status.client
}

// Sets the status as available
// clipModelDimensions - Dimensions of the CLIP model
// ranInitialScan - True if the initial scan was executed
func (s *SemanticSearchSystem) SetStatusAvailable(dimensions uint, client sse_proto.SemanticSearchEngineServiceClient) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	s.status.available = true
	s.status.dimensions = dimensions
	s.status.client = client
}

// Gets the image size limit for the CLIP encoder
func (s *SemanticSearchSystem) GetClipImageSizeLimit() int64 {
	return int64(s.imageSizeLimit)
}

const SSE_DB_PASSPHRASE_SALT = "semantic-search-db-key"

func makeDatabasePassPhrase(key []byte) string {
	hasher := sha256.New()
	hasher.Write(key)
	hasher.Write([]byte(SSE_DB_PASSPHRASE_SALT))
	hash := hasher.Sum(nil)
	return strings.ToLower(hex.EncodeToString(hash))
}

func readAndLogSemanticSearchEngineLogs(pipe io.ReadCloser) {
	reader := bufio.NewReader(pipe)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			return
		}

		LogLine("[SemanticSearch] [Engine] " + strings.TrimSpace(line))
	}
}

// Initializes engine (requires vault key)
func (s *SemanticSearchSystem) initializeEngine(key []byte) {
	dbPassphrase := makeDatabasePassPhrase(key)

	logLevel := "INFO"

	if log_debug_enabled {
		logLevel = "DEBUG"
	}

	cmd := exec.Command(s.sseBinPath, "-l", logLevel, s.modelPath, s.dbPath)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "API_KEY="+s.apiKey, "SQLITE_CIPHER_PASSPHRASE="+dbPassphrase)

	// Configure command
	err := child_process_manager.ConfigureCommand(cmd)
	if err != nil {
		LogErrorMsg("[SemanticSearch] [Error] Could not initialize engine: " + err.Error() + " | While preparing SSE engine server")
		return
	}

	// Create a pipe to read StdOut
	pipeOut, err := cmd.StdoutPipe()

	if err != nil {
		LogErrorMsg("[SemanticSearch] [Error] Could not initialize engine: " + err.Error() + " | While preparing SSE engine server")
		return
	}

	// Create a pipe to read StdErr
	pipeErr, err := cmd.StderrPipe()

	if err != nil {
		LogErrorMsg("[SemanticSearch] [Error] Could not initialize engine: " + err.Error() + " | While preparing SSE engine server")
		return
	}

	// Start the command

	LogDebug("Running command: " + cmd.String())

	err = cmd.Start()

	if err != nil {
		LogErrorMsg("[SemanticSearch] [Error] Could not initialize engine: " + err.Error() + " | While starting SSE engine server")
		return
	}

	// Add process as a child process
	child_process_manager.AddChildProcess(cmd.Process) //nolint:errcheck

	// Pipe stdErr to logs
	go readAndLogSemanticSearchEngineLogs(pipeErr)

	// Wait for stdOut port output

	outReader := bufio.NewReader(pipeOut)

	outLine, err := outReader.ReadString('\n')

	if err != nil {
		LogErrorMsg("[SemanticSearch] [Error] Could not initialize engine: " + err.Error() + " | While waiting for SSE server to be available")
		return
	}

	portStr := strings.TrimSpace(outLine)

	port, err := strconv.Atoi(portStr)

	if err != nil {
		LogErrorMsg("[SemanticSearch] [Error] Could not initialize engine: " + err.Error() + " | The SSE engine server returned an invalid port")
		return
	}

	// Create GRPC client

	conn, err := grpc.NewClient("127.0.0.1:"+fmt.Sprint(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		LogErrorMsg("[SemanticSearch] [Error] Could not initialize engine: " + err.Error() + " | While connecting to the GRPC server at port: " + fmt.Sprint(port))
		return
	}

	grpcClient := sse_proto.NewSemanticSearchEngineServiceClient(conn)

	// Get metadata

	metadataRes, err := grpcClient.GetModelMetadata(context.Background(), &sse_proto.ClipModelMetadataRequest{
		ApiKey: s.apiKey,
	})

	if err != nil {
		LogErrorMsg("[SemanticSearch] [Error] Could not initialize engine: " + err.Error() + " | While requesting metadata from the SSE engine server")
		return
	}

	dimensions := metadataRes.EmbedDim

	// Ready

	s.SetStatusAvailable(uint(dimensions), grpcClient)

	LogDebug("[SemanticSearch] Initialization successful. Service available.")

	// Initial scan

	go s.DoInitialScan(key)

	// Wait for the process

	err = cmd.Wait()

	if err != nil {
		LogErrorMsg("[Semantic Search] [Error] SSE engine server crashed: " + err.Error())
	}
}

type ClipEncodeTextRequest struct {
	Text string `json:"text"`
}

type ClipVectorResponse struct {
	Features []float32 `json:"features"`
}

func (s *SemanticSearchSystem) clipEncodeTextInternal(ctx context.Context, text string) ([]float32, error) {
	client := s.GetClient()

	if client == nil {
		return nil, errors.New("grpc client is not available")
	}

	response, err := client.EncodeText(ctx, &sse_proto.ClipTextEmbeddingRequest{
		ApiKey: s.apiKey,
		Text:   text,
	})

	if err != nil {
		return nil, err
	}

	if response.Features == nil {
		return nil, errors.New("received nil features in response")
	}

	return response.Features, nil
}

// Encodes text into a vector
func (s *SemanticSearchSystem) ClipEncodeText(ctx context.Context, text string) ([]float32, error) {
	return s.clipEncodeTextInternal(ctx, text)
}

const SSE_IMAGE_CHUNK_SIZE = 2 * 1024 * 1024

func (s *SemanticSearchSystem) clipEncodeImageInternal(ctx context.Context, image []byte) ([]float32, bool, error) {
	client := s.GetClient()

	if client == nil {
		return nil, false, errors.New("grpc client is not available")
	}

	streamingClient, err := client.EncodeImage(ctx)

	if err != nil {
		return nil, false, err
	}

	// Authenticate
	err = streamingClient.Send(&sse_proto.ClipImageEmbeddingRequest{
		ImageEmbeddingOneof: &sse_proto.ClipImageEmbeddingRequest_Init{
			Init: &sse_proto.ClipImageEmbeddingRequestInit{
				ApiKey:   s.apiKey,
				MimeType: "image/any",
			},
		},
	})

	if err != nil {
		_, _ = streamingClient.CloseAndRecv()
		return nil, false, err
	}

	// Send image chunks

	offset := 0

	for offset < len(image) {
		remainingBytes := len(image) - offset
		chunkSize := min(remainingBytes, SSE_IMAGE_CHUNK_SIZE)

		chunk := image[offset : offset+chunkSize]

		err = streamingClient.Send(&sse_proto.ClipImageEmbeddingRequest{
			ImageEmbeddingOneof: &sse_proto.ClipImageEmbeddingRequest_Chunk{
				Chunk: &sse_proto.ClipImageEmbeddingRequestChunk{
					ImageChunk: chunk,
				},
			},
		})

		if err != nil {
			_, _ = streamingClient.CloseAndRecv()
			return nil, false, err
		}

		offset += chunkSize
	}

	response, err := streamingClient.CloseAndRecv()

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			// Now you can check the specific codes
			switch st.Code() {
			case codes.InvalidArgument:
				return nil, true, err
			default:
				return nil, false, err
			}
		} else {
			return nil, false, err
		}
	}

	if response.Features == nil {
		return nil, false, errors.New("received nil features in response")
	}

	return response.Features, false, nil
}

// Encodes image into a vector
// image - Bytes of the image file
// Note: Make sure the file is not too big
// The file must be validated before calling this function
func (s *SemanticSearchSystem) ClipEncodeImage(ctx context.Context, image []byte) (vector []float32, isInvalidImageError bool, err error) {
	return s.clipEncodeImageInternal(ctx, image)
}

// Indexed vector
type SemanticSearchIndexedVector struct {
	// The ID of the vector
	Id uint64
	// The media ID
	Media uint64
	// A hash of the data
	DataHash string
	// The vector
	Vector []float32
}

func NewSemanticSearchIndexedVector(vector []float32, media_id uint64, dataHash string) (*SemanticSearchIndexedVector, error) {
	return &SemanticSearchIndexedVector{
		Media:    media_id,
		DataHash: dataHash,
		Vector:   vector,
	}, nil
}

// Finds all the indexed vectors for a specific media
func (s *SemanticSearchSystem) GetIndexedVectors(ctx context.Context, media uint64) ([]*SemanticSearchIndexedVector, error) {
	client := s.GetClient()

	if client == nil {
		return nil, errors.New("grpc client is not available")
	}

	response, err := client.GetVectorsByMedia(ctx, &sse_proto.GetVectorsByMediaRequest{
		ApiKey:  s.apiKey,
		MediaId: media,
	})

	if err != nil {
		return nil, err
	}

	result := make([]*SemanticSearchIndexedVector, len(response.Vectors))

	for i, vector := range response.Vectors {
		result[i] = &SemanticSearchIndexedVector{
			Id:       vector.VectorId,
			Media:    vector.MediaId,
			DataHash: vector.DataHash,
		}
	}

	return result, nil
}

// Semantic search query
type SemanticSearchQuery struct {
	// The vector
	Vector []float32

	// Max number of results to get
	Limit uint64

	// Continuation token
	ContinuationToken *float32
}

// Performs a vector query to the vector database
func (s *SemanticSearchSystem) QueryVectors(ctx context.Context, query *SemanticSearchQuery) ([]*SemanticSearchIndexedVector, *float32, error) {
	client := s.GetClient()

	if client == nil {
		return nil, nil, errors.New("grpc client is not available")
	}

	req := sse_proto.QueryVectorsRequest{
		ApiKey:            s.apiKey,
		Features:          query.Vector,
		Limit:             query.Limit,
		ContinuationToken: query.ContinuationToken,
	}

	response, err := client.QueryVectors(ctx, &req)

	if err != nil {
		return nil, nil, err
	}

	result := make([]*SemanticSearchIndexedVector, len(response.Vectors))

	for i, vector := range response.Vectors {
		result[i] = &SemanticSearchIndexedVector{
			Id:       vector.VectorId,
			Media:    vector.MediaId,
			DataHash: vector.DataHash,
		}
	}

	return result, response.ContinuationToken, nil
}

// Deletes vectors by IDs
func (s *SemanticSearchSystem) DeleteVectors(ctx context.Context, vectors []*SemanticSearchIndexedVector) error {
	if len(vectors) == 0 {
		return nil
	}

	client := s.GetClient()

	if client == nil {
		return errors.New("grpc client is not available")
	}

	vectorIds := make([]uint64, len(vectors))

	for i := range vectors {
		vectorIds[i] = vectors[i].Id
	}

	_, err := client.DeleteVectors(ctx, &sse_proto.DeleteVectorsRequest{
		ApiKey:    s.apiKey,
		VectorIds: vectorIds,
	})

	return err
}

// Inserts vectors into the database
// ctx - The execution context
// vectors - List of vectors to insert. make sure all vectors contain a non-nil 'Vector' field
func (s *SemanticSearchSystem) InsertVectors(ctx context.Context, vectors []*SemanticSearchIndexedVector) error {
	if len(vectors) == 0 {
		return nil
	}

	client := s.GetClient()

	if client == nil {
		return errors.New("grpc client is not available")
	}

	vectorsToInsert := make([]*sse_proto.InsertVectorRequest, len(vectors))

	for i, v := range vectors {
		vectorsToInsert[i] = &sse_proto.InsertVectorRequest{
			MediaId:  v.Media,
			DataHash: v.DataHash,
			Features: v.Vector,
		}
	}

	_, err := client.InsertVectors(ctx, &sse_proto.InsertVectorsRequest{
		ApiKey:   s.apiKey,
		Requests: vectorsToInsert,
	})

	return err
}

func (s *SemanticSearchSystem) removeMediaFromIndex(media_id uint64) {

	vectors, err := s.GetIndexedVectors(context.Background(), media_id)

	if err != nil {
		LogErrorMsg("Error fetching indexed vectors: " + err.Error())
		return
	}

	err = s.DeleteVectors(context.Background(), vectors)

	if err != nil {
		LogErrorMsg("Error deleting vectors: " + err.Error())
	}
}

func (s *SemanticSearchSystem) extractImageFromMedia(media_id uint64, original_asset uint64, key []byte) (image []byte, err error) {
	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		return nil, nil
	}

	found, asset_path, asset_lock := media.AcquireAsset(original_asset, ASSET_SINGLE_FILE)

	if !found {
		GetVault().media.ReleaseMediaResource(media_id)
		return nil, nil
	}

	asset_lock.StartRead() // Start reading the asset

	rs, err := encrypted_storage.CreateFileBlockEncryptReadStream(asset_path, key, FILE_PERMISSION)

	if err != nil {
		asset_lock.EndRead()
		media.ReleaseAsset(original_asset)
		GetVault().media.ReleaseMediaResource(media_id)

		return nil, errors.New("error reading asset file (" + asset_path + "): " + err.Error())
	}

	sizeLimit := int64(s.imageSizeLimit)

	if rs.FileSize() <= sizeLimit {
		imageData, err := io.ReadAll(rs)

		rs.Close()
		asset_lock.EndRead()
		media.ReleaseAsset(original_asset)
		GetVault().media.ReleaseMediaResource(media_id)

		return imageData, err
	}

	// Image too big, decrypt into a temporal file for encoding

	tempFile := GetTemporalFileName("png", false)

	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FILE_PERMISSION)

	if err != nil {
		rs.Close()
		asset_lock.EndRead()
		media.ReleaseAsset(original_asset)
		GetVault().media.ReleaseMediaResource(media_id)

		return nil, err
	}

	buf := make([]byte, 1024*1024)
	var finished = false

	for !finished {
		c, err := rs.Read(buf)

		if err != nil && err != io.EOF {
			_ = f.Close()
			DeleteTemporalFile(tempFile)

			rs.Close()
			asset_lock.EndRead()
			media.ReleaseAsset(original_asset)
			GetVault().media.ReleaseMediaResource(media_id)

			return nil, err
		}

		if err == io.EOF {
			finished = true
		}

		if c == 0 {
			continue
		}

		_, err = f.Write(buf[:c])

		if err != nil {
			_ = f.Close()
			DeleteTemporalFile(tempFile)

			rs.Close()
			asset_lock.EndRead()
			media.ReleaseAsset(original_asset)
			GetVault().media.ReleaseMediaResource(media_id)

			return nil, err
		}
	}

	_ = f.Close()
	rs.Close()

	asset_lock.EndRead()
	media.ReleaseAsset(original_asset)
	GetVault().media.ReleaseMediaResource(media_id)

	// Probe the image

	probe_data, err := ProbeMediaFileWithFFProbe(tempFile)

	if err != nil {
		DeleteTemporalFile(tempFile)

		return nil, err
	}

	if probe_data.Type != MediaTypeImage {
		DeleteTemporalFile(tempFile)

		return nil, nil
	}

	// Encode image to reduce its size

	userConfig, err := GetVault().config.Read(key)

	if err != nil {
		DeleteTemporalFile(tempFile)

		return nil, err
	}

	tempFolder, err := GetTemporalFolder(false)

	if err != nil {
		DeleteTemporalFile(tempFile)

		return nil, err
	}

	cmd := MakeFFMpegEncodeToPNGCommand(tempFile, probe_data.Format, tempFolder, &UserConfigResolution{
		Width:  900,
		Height: 900,
		Fps:    -1,
	}, probe_data.Width, probe_data.Height, userConfig)

	err = cmd.Run()

	if err != nil {
		DeleteTemporalFile(tempFile)
		DeleteTemporalPath(tempFolder)

		return nil, err
	}

	DeleteTemporalFile(tempFile)

	// Load the smaller image file into memory

	imageData, err := os.ReadFile(path.Join(tempFolder, "image.png"))

	DeleteTemporalPath(tempFolder)

	if err != nil {
		return nil, err
	}

	return imageData, nil
}

func (s *SemanticSearchSystem) addOrUpdateMediaIndex(media_id uint64, key []byte) {
	vectors, err := s.GetIndexedVectors(context.Background(), media_id)

	if err != nil {
		LogErrorMsg("Error fetching indexed vectors: " + err.Error())
		return
	}

	LogDebug("Found " + fmt.Sprint(len(vectors)) + " vector for media #" + fmt.Sprint(media_id))

	imageHash := ""

	for _, v := range vectors {
		imageHash = v.DataHash
	}

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		err = s.DeleteVectors(context.Background(), vectors)

		if err != nil {
			LogErrorMsg("Error deleting vectors: " + err.Error())
		}

		return
	}

	meta, err := media.ReadMetadata(key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	if meta == nil {
		err = s.DeleteVectors(context.Background(), vectors)

		if err != nil {
			LogErrorMsg("Error deleting vectors: " + err.Error())
		}

		return
	}

	vectorsToInsert := make([]*SemanticSearchIndexedVector, 0)

	// Image

	actualImageHash := fmt.Sprint(meta.OriginalAsset)

	if actualImageHash != imageHash || len(vectors) != 1 {
		// Re-index of image vector required

		if log_debug_enabled && actualImageHash != imageHash {
			LogDebug("Data hash mismatch (" + actualImageHash + " != " + imageHash + ")")
		}

		err = s.DeleteVectors(context.Background(), vectors)

		if err != nil {
			LogErrorMsg("Error deleting vectors: " + err.Error())
			return
		}

		if meta.Type == MediaTypeImage && meta.OriginalEncoded {
			image, err := s.extractImageFromMedia(media_id, meta.OriginalAsset, key)

			if err != nil {
				LogError(err)
			}

			if image != nil {
				features, isInvalidImageError, err := s.ClipEncodeImage(context.Background(), image)

				if isInvalidImageError {
					LogErrorMsg("Invalid image when encoding for indexing. media_id=" + fmt.Sprint(media_id) + ", asset_id=" + fmt.Sprint(meta.OriginalAsset))
					return
				} else if err != nil {
					LogErrorMsg("Error encoding image: " + err.Error())
					return
				} else {
					vectorImage, err := NewSemanticSearchIndexedVector(features, media_id, actualImageHash)

					if err != nil {
						LogErrorMsg("Error creating vector: " + err.Error())
						return
					}

					vectorsToInsert = append(vectorsToInsert, vectorImage)
				}
			}
		}
	}

	// Insert vectors

	err = s.InsertVectors(context.Background(), vectorsToInsert)

	if err != nil {
		LogErrorMsg("Error inserting vectors: " + err.Error())
	}
}

func (s *SemanticSearchSystem) doIndexingOperation(media_id uint64, isDeletion bool, key []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	finished := false

	for !finished {
		if isDeletion {
			s.removeMediaFromIndex(media_id)
			finished = true
		} else {
			s.addOrUpdateMediaIndex(media_id, key)
		}

		s.pendingMu.Lock()

		if finished {
			delete(s.pendingDelete, media_id)
			delete(s.pendingIndex, media_id)
			delete(s.busy, media_id)
		} else if s.pendingDelete[media_id] {
			delete(s.pendingDelete, media_id)
			isDeletion = true
		} else if s.pendingIndex[media_id] {
			delete(s.pendingIndex, media_id)
			isDeletion = false
		} else {
			delete(s.busy, media_id)
			finished = true
		}

		s.pendingMu.Unlock()
	}
}

// Request for the vectors associated with a media asset to be deleted
// from the vector database
func (s *SemanticSearchSystem) RequestMediaIndexRemoval(media_id uint64, key []byte, wait bool) {
	s.pendingMu.Lock()

	waitGroup := s.busy[media_id]

	canStartOperation := waitGroup == nil

	if waitGroup != nil {
		delete(s.pendingIndex, media_id)

		s.pendingDelete[media_id] = true
	} else {
		waitGroup = &sync.WaitGroup{}
		waitGroup.Add(1)
		s.busy[media_id] = waitGroup
	}

	s.pendingMu.Unlock()

	if canStartOperation {
		go s.doIndexingOperation(media_id, true, key, waitGroup)
	}

	if wait {
		waitGroup.Wait()
	}
}

func (s *SemanticSearchSystem) RequestMediaIndexing(media_id uint64, key []byte, wait bool) {
	s.pendingMu.Lock()

	waitGroup := s.busy[media_id]

	canStartOperation := waitGroup == nil

	if waitGroup != nil {
		if !s.pendingDelete[media_id] {
			s.pendingIndex[media_id] = true
		}
	} else {
		waitGroup = &sync.WaitGroup{}
		waitGroup.Add(1)
		s.busy[media_id] = waitGroup
	}

	s.pendingMu.Unlock()

	if canStartOperation {
		go s.doIndexingOperation(media_id, false, key, waitGroup)
	}

	if wait {
		waitGroup.Wait()
	}
}

const SSE_INITIAL_SCAN_PAGE_SIZE = 64

func (s *SemanticSearchSystem) getInitialScanPage(page int64) (items []uint64, isEnd bool, err error) {
	skip := page * SSE_INITIAL_SCAN_PAGE_SIZE

	main_index, err := GetVault().index.StartRead()

	if err != nil {
		return nil, false, err
	}

	defer GetVault().index.EndRead(main_index)

	page_items, err := main_index.ListValues(skip, SSE_INITIAL_SCAN_PAGE_SIZE)

	if err != nil {
		return nil, false, err
	}

	return page_items, len(page_items) >= SSE_INITIAL_SCAN_PAGE_SIZE, nil
}

// Run the initial scan
func (s *SemanticSearchSystem) DoInitialScan(key []byte) {
	finished := false

	page := int64(0)

	for !finished {
		pageItems, isEnd, err := s.getInitialScanPage(page)

		if err != nil {
			LogError(err)
			return
		}

		for _, mediaId := range pageItems {
			LogDebug("[SemanticSearch] [INITIAL SCAN] Checking media #" + fmt.Sprint(mediaId))
			s.RequestMediaIndexing(mediaId, key, true)
		}

		page++

		finished = !isEnd
	}
}

// Tries to set the initialized status
func (s *SemanticSearchSystem) TrySetInitialized() bool {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	if s.status.initialized {
		return false
	}

	s.status.initialized = true

	return true
}

func (s *SemanticSearchSystem) OnNewSession(session *ActiveSession) {
	mustRun := s.TrySetInitialized()

	if mustRun {
		go s.initializeEngine(session.key)
	}
}
