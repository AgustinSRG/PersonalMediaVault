// Semantic search system

// cSpell:ignore uuid, NewIDUUID

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

const (
	QDRANT_FIELD_MEDIA       = "m"
	QDRANT_FIELD_VECTOR_TYPE = "t"
	QDRANT_FIELD_DATA_HASH   = "h"
)

// Semantic search configuration
type SemanticSearchConfig struct {
	Enabled            bool
	QdrantHost         string
	QdrantPort         int
	QdrantApiKey       string
	QdrantInitialScan  bool
	ClipApiBaseUrl     string
	ClipApiAuth        string
	ClipImageSizeLimit uint64
}

const QDRANT_DEFAULT_PORT = 6334
const CLIP_DEFAULT_SIZE_LIMIT_MB = 20

func LoadSemanticSearchConfig() *SemanticSearchConfig {
	if os.Getenv("SEMANTIC_SEARCH_ENABLED") != "YES" {
		return &SemanticSearchConfig{
			Enabled: false,
		}
	}

	qDrantHost := os.Getenv("QDRANT_HOST")

	if qDrantHost == "" {
		LogWarning("QDRANT_HOST is unset. Using 'localhost' as default.")
		qDrantHost = "localhost"
	}

	qdrantPortStr := os.Getenv("QDRANT_PORT")
	qdrantPort := QDRANT_DEFAULT_PORT

	p, err := strconv.Atoi(qdrantPortStr)

	if err == nil {
		qdrantPort = p
	} else {
		LogWarning("Error parsing QDRANT_PORT value: " + err.Error() + " | Using " + fmt.Sprint(qdrantPort) + " as fallback value.")
	}

	qdrantApiKey := os.Getenv("QDRANT_API_KEY")

	if qdrantApiKey == "" {
		LogWarning("QDRANT_API_KEY is empty. This will probably cause authentication errors when connecting to the Qdrant database.")
	}

	qdrantInitialScan := strings.ToUpper(os.Getenv("QDRANT_INITIAL_SCAN")) != "NO"

	clipApiBase := os.Getenv("CLIP_API_BASE")

	if clipApiBase != "" {
		clipApiBase = "http://localhost:5000/clip"
		LogWarning("CLIP_API_BASE is empty. Using '" + clipApiBase + "' as the default value.")
	}

	clipApiAuth := os.Getenv("CLIP_API_AUTH")

	if clipApiAuth == "" {
		LogWarning("CLIP_API_AUTH is empty. This will probably cause authentication errors when calling the CLIP API.")
	}

	clipImageSizeLimit := uint64(CLIP_DEFAULT_SIZE_LIMIT_MB) * 1024 * 1024
	clipImageSizeLimitMbStr := os.Getenv("CLIP_IMAGE_SIZE_LIMIT_MB")

	if clipImageSizeLimitMbStr != "" {
		v, err := strconv.ParseUint(clipImageSizeLimitMbStr, 10, 32)

		if err == nil {
			clipImageSizeLimit = v * 1024 * 1024
		} else {
			LogWarning("Error parsing CLIP_IMAGE_SIZE_LIMIT_MB value: " + err.Error() + " | Using " + fmt.Sprint(CLIP_DEFAULT_SIZE_LIMIT_MB) + " as fallback value.")
		}
	}

	return &SemanticSearchConfig{
		Enabled:            true,
		QdrantHost:         qDrantHost,
		QdrantPort:         qdrantPort,
		QdrantApiKey:       qdrantApiKey,
		QdrantInitialScan:  qdrantInitialScan,
		ClipApiBaseUrl:     clipApiBase,
		ClipApiAuth:        clipApiAuth,
		ClipImageSizeLimit: clipImageSizeLimit,
	}
}

// Status for SemanticSearchSystem
type SemanticSearchSystemStatus struct {
	// Is the service available?
	available bool

	// Dimensions for the
	clipModelDimensions uint
}

// Semantic search sub-system
type SemanticSearchSystem struct {
	// Qdrant client
	qdrantClient *qdrant.Client

	// Configuration
	qDrantCollectionName string
	qdrantInitialScan    bool
	clipBaseUrl          string
	clipEncodeTextUrl    string
	clipEncodeImageUrl   string
	clipApiAuth          string
	clipImageSizeLimit   int64

	// Status
	status   SemanticSearchSystemStatus
	statusMu *sync.Mutex

	// Qdrant pending state
	qDrantBusy          map[uint64]*sync.WaitGroup
	qDrantPendingIndex  map[uint64]bool
	qDrantPendingDelete map[uint64]bool
	qDrantPendingMu     *sync.Mutex
}

// Creates instance of SemanticSearchSystem
func CreateSemanticSearchSystem(config *SemanticSearchConfig, vaultFingerprint string) (*SemanticSearchSystem, error) {
	qdrantClient, err := qdrant.NewClient(&qdrant.Config{
		Host:   config.QdrantHost,
		Port:   config.QdrantPort,
		APIKey: config.QdrantApiKey,
	})

	if err != nil {
		return nil, err
	}

	clipEncodeTextUrl := config.ClipApiBaseUrl + "/encode/text"
	clipEncodeImageUrl := config.ClipApiBaseUrl + "/encode/image"

	return &SemanticSearchSystem{
		qdrantClient: qdrantClient,

		qDrantCollectionName: "pmv_" + vaultFingerprint,
		qdrantInitialScan:    config.QdrantInitialScan,
		clipBaseUrl:          config.ClipApiBaseUrl,
		clipEncodeTextUrl:    clipEncodeTextUrl,
		clipEncodeImageUrl:   clipEncodeImageUrl,
		clipApiAuth:          config.ClipApiAuth,
		clipImageSizeLimit:   int64(config.ClipImageSizeLimit),

		status: SemanticSearchSystemStatus{
			available:           false,
			clipModelDimensions: 0,
		},
		statusMu: &sync.Mutex{},

		qDrantBusy:          make(map[uint64]*sync.WaitGroup),
		qDrantPendingIndex:  make(map[uint64]bool),
		qDrantPendingDelete: make(map[uint64]bool),
		qDrantPendingMu:     &sync.Mutex{},
	}, nil
}

// Gets the status of the sub-system
func (s *SemanticSearchSystem) GetStatus() SemanticSearchSystemStatus {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	return s.status
}

// Sets the status as available
// clipModelDimensions - Dimensions of the CLIP model
func (s *SemanticSearchSystem) SetStatusAvailable(clipModelDimensions uint) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()

	s.status.available = true
	s.status.clipModelDimensions = clipModelDimensions
}

// Gets the image size limit for the CLIP encoder
func (s *SemanticSearchSystem) GetClipImageSizeLimit() int64 {
	return s.clipImageSizeLimit
}

type ClipModelMetadata struct {
	Name       string `json:"name"`
	Dimensions uint   `json:"dimensions"`
}

type ClipApiMetadataResponse struct {
	Name   string            `json:"name"`
	Device string            `json:"device"`
	Model  ClipModelMetadata `json:"model"`
}

func (s *SemanticSearchSystem) getClipModelDimensions() (uint, error) {
	resp, err := http.Get(s.clipBaseUrl)

	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, errors.New("not successful status code: " + fmt.Sprint(resp.StatusCode))
	}

	var p ClipApiMetadataResponse

	err = json.NewDecoder(resp.Body).Decode(&p)

	if err != nil {
		return 0, err
	}

	return p.Model.Dimensions, nil
}

func (s *SemanticSearchSystem) Initialize() {
	go s.initializeInternal()
}

func (s *SemanticSearchSystem) initializeInternal() {
	// Check model

	doneCheckingModel := false
	modelDimensions := uint(0)

	for !doneCheckingModel {
		dimensions, err := s.getClipModelDimensions()

		if err != nil {
			LogErrorMsg("[SemanticSearch] Error fetching CLIP model metadata: " + err.Error() + ". Trying again in 2 seconds...")
			time.Sleep(2 * time.Second)
			continue
		}

		modelDimensions = dimensions
		doneCheckingModel = true

		LogInfo("[SemanticSearch] Loaded CLIP model. Dimensions: " + fmt.Sprint(modelDimensions))
	}

	// Check database

	doneCheckingDatabase := false

	for !doneCheckingDatabase {
		exists, err := s.qdrantClient.CollectionExists(context.Background(), s.qDrantCollectionName)

		if err != nil {
			LogErrorMsg("[SemanticSearch] Error checking Qdrant database: " + err.Error() + ". Trying again in 2 seconds...")
			time.Sleep(2 * time.Second)
			continue
		}

		if exists {
			// Check if dimensions match the model

			collectionInfo, err := s.qdrantClient.GetCollectionInfo(context.Background(), s.qDrantCollectionName)

			if err != nil {
				LogErrorMsg("[SemanticSearch] Error checking Qdrant database info: " + err.Error() + ". Trying again in 2 seconds...")
				time.Sleep(2 * time.Second)
				continue
			}

			qdrantVectorSize := collectionInfo.Config.Params.VectorsConfig.GetParams().Size

			if uint64(modelDimensions) != qdrantVectorSize {
				LogWarning("[SemanticSearch] Qdrant database vector size (" +
					fmt.Sprint(qdrantVectorSize) + ") does not match with the CLIP model (" +
					fmt.Sprint(modelDimensions) + "). Deleting the database to create a new one...")

				err = s.qdrantClient.DeleteCollection(context.Background(), s.qDrantCollectionName)

				if err != nil {
					LogErrorMsg("[SemanticSearch] Error deleting Qdrant database: " + err.Error() + ". Trying again in 2 seconds...")
					time.Sleep(2 * time.Second)
				}

				continue
			}

			LogDebug("[SemanticSearch] Qdrant collection: " + s.qDrantCollectionName)
		} else {
			// Create

			err = s.qdrantClient.CreateCollection(context.Background(), &qdrant.CreateCollection{
				CollectionName: s.qDrantCollectionName,
				VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
					Size:     uint64(modelDimensions),
					Distance: qdrant.Distance_Cosine,
				}),
			})

			if err != nil {
				LogErrorMsg("[SemanticSearch] Error creating Qdrant database: " + err.Error() + ". Trying again in 2 seconds...")
				time.Sleep(2 * time.Second)
				continue
			}

			doneCreatingIndexes := false

			for !doneCreatingIndexes {
				mediaFieldType := qdrant.FieldType_FieldTypeInteger

				_, err := s.qdrantClient.CreateFieldIndex(context.Background(), &qdrant.CreateFieldIndexCollection{
					CollectionName: s.qDrantCollectionName,
					FieldName:      QDRANT_FIELD_MEDIA,
					FieldType:      &mediaFieldType,
				})

				if err != nil {
					LogErrorMsg("[SemanticSearch] Error creating Qdrant index: " + err.Error() + ". Trying again in 2 seconds...")
					time.Sleep(2 * time.Second)
					continue
				}

				doneCreatingIndexes = true
			}

			LogInfo("[SemanticSearch] Qdrant collection created: " + s.qDrantCollectionName)
		}

		doneCheckingDatabase = true
	}

	// Done

	s.SetStatusAvailable(modelDimensions)

	LogDebug("[SemanticSearch] Initialization successful. Service available.")
}

type ClipEncodeTextRequest struct {
	Text string `json:"text"`
}

type ClipVectorResponse struct {
	Features []float32 `json:"features"`
}

func (s *SemanticSearchSystem) clipEncodeTextInternal(text string) ([]float32, error) {
	body := ClipEncodeTextRequest{
		Text: text,
	}

	jsonRes, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	resp, err := http.Post(s.clipEncodeTextUrl, "application/json", bytes.NewReader(jsonRes))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("not successful status code: " + fmt.Sprint(resp.StatusCode))
	}

	var p ClipVectorResponse

	err = json.NewDecoder(resp.Body).Decode(&p)

	if err != nil {
		return nil, err
	}

	if p.Features == nil {
		return nil, errors.New("received nil features")
	}

	return p.Features, nil
}

// Encodes text into a vector
func (s *SemanticSearchSystem) ClipEncodeText(text string) ([]float32, error) {
	return s.clipEncodeTextInternal(text)
}

func (s *SemanticSearchSystem) clipEncodeImageInternal(image []byte) ([]float32, bool, error) {
	resp, err := http.Post(s.clipEncodeImageUrl, "application/json", bytes.NewReader(image))

	if err != nil {
		return nil, false, err
	}

	if resp.StatusCode != 200 {
		return nil, resp.StatusCode == 400, errors.New("not successful status code: " + fmt.Sprint(resp.StatusCode))
	}

	var p ClipVectorResponse

	err = json.NewDecoder(resp.Body).Decode(&p)

	if err != nil {
		return nil, false, err
	}

	if p.Features == nil {
		return nil, false, errors.New("received nil features")
	}

	return p.Features, false, nil
}

// Encodes image into a vector
// image - Bytes of the image file
// Note: Make sure the file is not too big
// The file must be validated before calling this function
func (s *SemanticSearchSystem) ClipEncodeImage(image []byte) (vector []float32, isInvalidImageError bool, err error) {
	return s.clipEncodeImageInternal(image)
}

// Indexed vector types
type QdrantIndexedVectorType int

const (
	VECTOR_TYPE_TEXT  QdrantIndexedVectorType = 0
	VECTOR_TYPE_IMAGE QdrantIndexedVectorType = 1
)

// Indexed vector
type QdrantIndexedVector struct {
	// The ID of the vector
	Id *qdrant.PointId
	// The media ID
	Media uint64
	// The type of vector
	VectorType QdrantIndexedVectorType
	// A hash of the data
	DataHash string
	// The vector
	Vector []float32
}

func NewQdrantIndexedVector(vector []float32, media_id uint64, vectorType QdrantIndexedVectorType, dataHash string) (*QdrantIndexedVector, error) {
	id, err := uuid.NewV7()

	if err != nil {
		return nil, err
	}

	return &QdrantIndexedVector{
		Id:         qdrant.NewIDUUID(id.String()),
		Media:      media_id,
		VectorType: vectorType,
		DataHash:   dataHash,
		Vector:     vector,
	}, nil
}

func QdrantIndexedVectorFromScoredPoint(p *qdrant.ScoredPoint) *QdrantIndexedVector {
	mediaId := uint64(0)
	vectorType := VECTOR_TYPE_TEXT
	dataHash := ""

	if p.Payload != nil {
		mediaIdVal := p.Payload[QDRANT_FIELD_MEDIA]

		if mediaIdVal != nil {
			mediaId = uint64(mediaIdVal.GetIntegerValue())
		}

		vectorTypeVal := p.Payload[QDRANT_FIELD_VECTOR_TYPE]

		if vectorTypeVal != nil {
			vectorType = QdrantIndexedVectorType(vectorTypeVal.GetIntegerValue())
		}

		dataHashVal := p.Payload[QDRANT_FIELD_DATA_HASH]

		if dataHashVal != nil {
			dataHash = dataHashVal.GetStringValue()
		}
	}

	var vector []float32 = nil

	if p.Vectors != nil {
		vo := p.Vectors.GetVector()

		if vo != nil {
			dense := vo.GetDense()

			if dense != nil {
				vector = dense.GetData()
			}
		}
	}

	return &QdrantIndexedVector{
		Id:         p.Id,
		Media:      mediaId,
		VectorType: vectorType,
		DataHash:   dataHash,
		Vector:     vector,
	}
}

// Finds all the indexed vectors for a specific media
func (s *SemanticSearchSystem) GetIndexedVectors(ctx context.Context, media uint64) ([]*QdrantIndexedVector, error) {
	queryLimit := uint64(10)

	searchResult, err := s.qdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: s.qDrantCollectionName,
		Query:          nil,
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatchInt(QDRANT_FIELD_MEDIA, int64(media)),
			},
		},
		WithPayload: qdrant.NewWithPayload(true),
		Limit:       &queryLimit,
	})

	if err != nil {
		return nil, err
	}

	result := make([]*QdrantIndexedVector, len(searchResult))

	for i := range searchResult {
		result[i] = QdrantIndexedVectorFromScoredPoint(searchResult[i])
	}

	return result, nil
}

// Semantic search query
type SemanticSearchQuery struct {
	// The vector
	Vector []float32

	// The vector type (optional)
	VectorType *QdrantIndexedVectorType

	// Number of results to skip
	Offset uint64

	// Max number of results to get
	Limit uint64
}

// Performs a vector query to the Qdrant database
func (s *SemanticSearchSystem) QueryVectors(ctx context.Context, query *SemanticSearchQuery) ([]*QdrantIndexedVector, error) {
	var queryFilter *qdrant.Filter = nil

	if query.VectorType != nil {
		queryFilter = &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatchInt(QDRANT_FIELD_VECTOR_TYPE, int64(*query.VectorType)),
			},
		}
	}

	queryOffset := &query.Offset

	if *queryOffset == 0 {
		queryOffset = nil
	}

	searchResult, err := s.qdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: s.qDrantCollectionName,
		Query:          qdrant.NewQuery(query.Vector...),
		Filter:         queryFilter,
		WithPayload:    qdrant.NewWithPayload(true),
		Limit:          &query.Limit,
		Offset:         queryOffset,
	})

	if err != nil {
		return nil, err
	}

	result := make([]*QdrantIndexedVector, len(searchResult))

	for i := range searchResult {
		result[i] = QdrantIndexedVectorFromScoredPoint(searchResult[i])
	}

	return result, nil
}

// Deletes vectors by IDs
func (s *SemanticSearchSystem) DeleteVectors(ctx context.Context, vectors []*QdrantIndexedVector) error {
	if len(vectors) == 0 {
		return nil
	}

	vectorIds := make([]*qdrant.PointId, len(vectors))

	for i := range vectors {
		vectorIds[i] = vectors[i].Id
	}

	_, err := s.qdrantClient.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: s.qDrantCollectionName,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: vectorIds,
				},
			},
		},
	})

	return err
}

// Inserts vectors into the Qdrant database
// ctx - The execution context
// vectors - List of vectors to insert. make sure all vectors contain a non-nil 'Vector' field
func (s *SemanticSearchSystem) InsertVectors(ctx context.Context, vectors []*QdrantIndexedVector) error {
	if len(vectors) == 0 {
		return nil
	}

	points := make([]*qdrant.PointStruct, len(vectors))

	for i, v := range vectors {
		points[i] = &qdrant.PointStruct{
			Id:      v.Id,
			Vectors: qdrant.NewVectors(v.Vector...),
			Payload: qdrant.NewValueMap(map[string]any{
				QDRANT_FIELD_MEDIA:       int64(v.Media),
				QDRANT_FIELD_VECTOR_TYPE: int64(v.VectorType),
				QDRANT_FIELD_DATA_HASH:   v.DataHash,
			}),
		}
	}

	_, err := s.qdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.qDrantCollectionName,
		Points:         points,
	})

	return err
}

func (s *SemanticSearchSystem) removeMediaFromQdrantIndex(media_id uint64) {
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

	sizeLimit := s.clipImageSizeLimit

	if rs.FileSize() > sizeLimit {
		rs.Close()
		asset_lock.EndRead()
		media.ReleaseAsset(original_asset)
		GetVault().media.ReleaseMediaResource(media_id)

		return nil, nil
	}

	imageData, err := io.ReadAll(rs)

	rs.Close()
	asset_lock.EndRead()
	media.ReleaseAsset(original_asset)
	GetVault().media.ReleaseMediaResource(media_id)

	return imageData, err

}

func (s *SemanticSearchSystem) addOrUpdateMediaIndex(media_id uint64, key []byte) {
	vectors, err := s.GetIndexedVectors(context.Background(), media_id)

	if err != nil {
		LogErrorMsg("Error fetching indexed vectors: " + err.Error())
		return
	}

	vectorsTitle := make([]*QdrantIndexedVector, 0)
	vectorsImage := make([]*QdrantIndexedVector, 0)

	titleHash := ""
	imageHash := ""

	for _, v := range vectors {
		switch v.VectorType {
		case VECTOR_TYPE_TEXT:
			vectorsTitle = append(vectorsTitle, v)
			titleHash = v.DataHash
		case VECTOR_TYPE_IMAGE:
			vectorsImage = append(vectorsImage, v)
			imageHash = v.DataHash
		}
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

	vectorsToInsert := make([]*QdrantIndexedVector, 0)

	// Title

	actualTitleHash := strings.ToLower(hex.EncodeToString(sha256.New().Sum([]byte(meta.Title))))

	if actualTitleHash != titleHash || len(vectorsTitle) != 1 {
		// Re-index of title vector required

		err = s.DeleteVectors(context.Background(), vectorsTitle)

		if err != nil {
			LogErrorMsg("Error deleting vectors: " + err.Error())
			return
		}

		if len(meta.Title) > 0 {
			features, err := s.ClipEncodeText(meta.Title)

			if err != nil {
				LogErrorMsg("Error encoding title: " + err.Error())
				return
			}

			vectorTitle, err := NewQdrantIndexedVector(features, media_id, VECTOR_TYPE_TEXT, actualTitleHash)

			if err != nil {
				LogErrorMsg("Error creating vector: " + err.Error())
				return
			}

			vectorsToInsert = append(vectorsToInsert, vectorTitle)
		}
	}

	// Image

	actualImageHash := fmt.Sprint(meta.OriginalAsset)

	if actualImageHash != imageHash || len(vectorsImage) != 1 {
		// Re-index of image vector required

		err = s.DeleteVectors(context.Background(), vectorsImage)

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
				features, isInvalidImageError, err := s.ClipEncodeImage(image)

				if isInvalidImageError {
					LogErrorMsg("Invalid image when encoding for indexing. media_id=" + fmt.Sprint(media_id) + ", asset_id=" + fmt.Sprint(meta.OriginalAsset))
					return
				} else if err != nil {
					LogErrorMsg("Error encoding image: " + err.Error())
					return
				} else {
					vectorImage, err := NewQdrantIndexedVector(features, media_id, VECTOR_TYPE_IMAGE, actualImageHash)

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

func (s *SemanticSearchSystem) doQdrantIndexingOperation(media_id uint64, isDeletion bool, key []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	finished := false

	for !finished {
		if isDeletion {
			s.removeMediaFromQdrantIndex(media_id)
			finished = true
		} else {
			s.addOrUpdateMediaIndex(media_id, key)
		}

		s.qDrantPendingMu.Lock()

		if finished {
			delete(s.qDrantPendingDelete, media_id)
			delete(s.qDrantPendingIndex, media_id)
			delete(s.qDrantBusy, media_id)
		} else if s.qDrantPendingDelete[media_id] {
			delete(s.qDrantPendingDelete, media_id)
			isDeletion = true
		} else if s.qDrantPendingIndex[media_id] {
			delete(s.qDrantPendingIndex, media_id)
			isDeletion = false
		} else {
			delete(s.qDrantBusy, media_id)
			finished = true
		}

		s.qDrantPendingMu.Unlock()
	}
}

// Request for the vectors associated with a media asset to be deleted
// from the Qdrant database
func (s *SemanticSearchSystem) RequestMediaIndexRemoval(media_id uint64, key []byte, wait bool) {
	s.qDrantPendingMu.Lock()

	waitGroup := s.qDrantBusy[media_id]

	canStartOperation := waitGroup == nil

	if waitGroup != nil {
		delete(s.qDrantPendingIndex, media_id)

		s.qDrantPendingDelete[media_id] = true
	} else {
		waitGroup = &sync.WaitGroup{}
		waitGroup.Add(1)
		s.qDrantBusy[media_id] = waitGroup
	}

	s.qDrantPendingMu.Unlock()

	if canStartOperation {
		go s.doQdrantIndexingOperation(media_id, true, key, waitGroup)
	}

	if wait {
		waitGroup.Wait()
	}
}

func (s *SemanticSearchSystem) RequestMediaIndexing(media_id uint64, key []byte, wait bool) {
	s.qDrantPendingMu.Lock()

	waitGroup := s.qDrantBusy[media_id]

	canStartOperation := waitGroup == nil

	if waitGroup != nil {
		if !s.qDrantPendingDelete[media_id] {
			s.qDrantPendingIndex[media_id] = true
		}
	} else {
		waitGroup = &sync.WaitGroup{}
		waitGroup.Add(1)
		s.qDrantBusy[media_id] = waitGroup
	}

	s.qDrantPendingMu.Unlock()

	if canStartOperation {
		go s.doQdrantIndexingOperation(media_id, false, key, waitGroup)
	}

	if wait {
		waitGroup.Wait()
	}
}
