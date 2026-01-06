package handlers

import (
	"encoding/json"
	"geo-feed-service/internal/ingestion"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type IngestionAPIUseCases interface {
	IngestData(input ingestion.IngestDataInput) error
}

type IngestionAPI struct {
	uc IngestionAPIUseCases
}

func NewIngestionAPI(uc IngestionAPIUseCases) *IngestionAPI {
	return &IngestionAPI{uc: uc}
}

type CreateContentRequest struct {
	// Optional external ID (publisher ID, event ID, etc.)
	ExternalID string `json:"external_id"`
	// Domain type: article, event, alert, etc.
	Type string `json:"type"`
	// Location as raw coordinates (not WKB yet)
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	// When this item should become visible to users
	PublishedAt string `json:"published_at"`
	// Arbitrary metadata (JSON)
	Attributes json.RawMessage `json:"attributes"`
	// Optional base relevance (defaults to 1.0)
	BaseScore float64 `json:"base_score"`
}

type CreateContentResponse struct {
	ID []byte `json:"id"`
}

func (api *IngestionAPI) CreateContent(writer http.ResponseWriter, req *http.Request) {
	slog.Info("ingestion started")
	var request CreateContentRequest

	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		slog.Error("Failed to unmarshal request body", "error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	input := api.toIngestDataInput(request)
	err = api.uc.IngestData(input)
	if err != nil {
		slog.Error("Failed to ingest data", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (api *IngestionAPI) toIngestDataInput(request CreateContentRequest) ingestion.IngestDataInput {
	return ingestion.IngestDataInput{
		ExternalID: []byte(request.ExternalID),
		Type:       []byte(request.Type),
		Lat:        request.Lat,
		Lng:        request.Lng,
		PublishedAt: func() time.Time {
			t, _ := time.Parse(time.RFC3339, request.PublishedAt)
			return t
		}(),
		Attributes: request.Attributes,
		BaseScore:  request.BaseScore,
	}
}
