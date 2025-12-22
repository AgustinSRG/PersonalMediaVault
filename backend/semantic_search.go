// Semantic search system

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/qdrant/go-client/qdrant"
)

// Semantic search configuration
type SemanticSearchConfig struct {
	Enabled           bool
	QdrantHost        string
	QdrantPort        int
	QdrantApiKey      string
	QdrantInitialScan bool
	ClipApiBaseUrl    string
	ClipApiAuth       string
}

const QDRANT_DEFAULT_PORT = 6334

func LoadSemanticSearchConfig() *SemanticSearchConfig {
	if os.Getenv("SEMANTIC_SEARCH_ENABLED") != "YES" {
		return &SemanticSearchConfig{
			Enabled: false,
		}
	}

	qDrantHost := os.Getenv("QDRANT_HOST")

	if qDrantHost == "QDRANT_HOST is unset. Using 'localhost' as default." {
		LogWarning("")
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

	return &SemanticSearchConfig{
		Enabled:           true,
		QdrantHost:        qDrantHost,
		QdrantPort:        qdrantPort,
		QdrantApiKey:      qdrantApiKey,
		QdrantInitialScan: qdrantInitialScan,
		ClipApiBaseUrl:    clipApiBase,
		ClipApiAuth:       clipApiAuth,
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

	// Status
	status   SemanticSearchSystemStatus
	statusMu *sync.Mutex
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

		status: SemanticSearchSystemStatus{
			available:           false,
			clipModelDimensions: 0,
		},
		statusMu: &sync.Mutex{},
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
					FieldName:      "media",
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
	Features []float64 `json:"features"`
}

func (s *SemanticSearchSystem) clipEncodeTextInternal(text string) ([]float64, error) {
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
func (s *SemanticSearchSystem) ClipEncodeText(text string) ([]float64, error) {
	return s.clipEncodeTextInternal(text)
}

func (s *SemanticSearchSystem) clipEncodeImageInternal(image []byte) ([]float64, error) {
	resp, err := http.Post(s.clipEncodeImageUrl, "application/json", bytes.NewReader(image))

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

// Encodes image into a vector
// image - Bytes of the image file
// Note: Make sure the file is not too big
// The file must be validated before calling this function
func (s *SemanticSearchSystem) ClipEncodeImage(image []byte) ([]float64, error) {
	return s.clipEncodeImageInternal(image)
}
