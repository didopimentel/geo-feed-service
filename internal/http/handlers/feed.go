package handlers

import (
	"context"
	"encoding/json"
	"geo-feed-service/internal/entities"
	"geo-feed-service/internal/feed"
	"io"
	"log/slog"
	"net/http"
)

type FeedAPIUseCases interface {
	GetFeed(ctx context.Context, input feed.GetFeedInput) (*entities.Feed, error)
}

type FeedAPI struct {
	uc FeedAPIUseCases
}

func NewFeedAPI(uc FeedAPIUseCases) *FeedAPI {
	return &FeedAPI{uc: uc}
}

type GetFeedRequest struct {
	// User location
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`

	// Search radius in meters
	RadiusMeters int `json:"radius_meters"`

	// Optional content type filters (article, event, alert, etc.)
	Types []string `json:"types"`

	// Max number of items to return
	// Defaults to 20, max capped server-side
	Limit int `json:"limit"`

	// Opaque cursor for pagination
	Cursor string `json:"cursor"`
}
type GetFeedResponse struct {
	Items      []entities.FeedItem `json:"items"`
	NextCursor string              `json:"next_cursor"`
}

func (api *FeedAPI) GetFeed(writer http.ResponseWriter, req *http.Request) {
	slog.Info("feed query started")
	var request GetFeedRequest

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

	input, err := toGetFeedInput(request)
	if err != nil {
		slog.Error("Invalid feed query", "error", err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	feedData, err := api.uc.GetFeed(req.Context(), input)
	if err != nil {
		slog.Error("Failed to get feed data", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	var nextCursor string
	if feedData.NextCursor != nil {
		nextCursor, err = entities.EncodeCursor(feedData.NextCursor)
		if err != nil {
			slog.Error("Failed to encode next cursor", "error", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	response := toGetFeedResponse(feedData, nextCursor)

	respBytes, err := json.Marshal(response)
	if err != nil {
		slog.Error("Failed to marshal response", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(respBytes)
}

func toGetFeedResponse(feedData *entities.Feed, nextCursor string) GetFeedResponse {
	return GetFeedResponse{
		Items:      feedData.Items,
		NextCursor: nextCursor,
	}
}

func toGetFeedInput(req GetFeedRequest) (feed.GetFeedInput, error) {
	var cursor *entities.Cursor
	if req.Cursor != "" {
		c, err := entities.DecodeCursor(req.Cursor)
		if err != nil {
			return feed.GetFeedInput{}, err
		}
		cursor = c
	}
	types := make([][]byte, 0, len(req.Types))
	for _, t := range req.Types {
		types = append(types, []byte(t))
	}
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return feed.GetFeedInput{
		Lat:          req.Lat,
		Lng:          req.Lng,
		RadiusMeters: req.RadiusMeters,
		Types:        types,
		Limit:        limit,
		Cursor:       cursor,
	}, nil
}
