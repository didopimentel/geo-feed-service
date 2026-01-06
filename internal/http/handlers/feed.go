package handlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"geo-feed-service/internal/entities"
	"geo-feed-service/internal/feed"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
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

// FeedItemDTO is the API representation of a feed item.
type FeedItemDTO struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Lat         float64         `json:"lat"`
	Lng         float64         `json:"lng"`
	PublishedAt time.Time       `json:"published_at"`
	Attributes  json.RawMessage `json:"attributes"`
	Score       float64         `json:"score"`
}
type GetFeedResponse struct {
	Items      []FeedItemDTO `json:"items"`
	NextCursor string        `json:"next_cursor"`
}

func (api *FeedAPI) GetFeed(writer http.ResponseWriter, req *http.Request) {
	slog.Info("feed query started")

	q := req.URL.Query()

	lat, err := strconv.ParseFloat(q.Get("lat"), 64)
	if err != nil {
		http.Error(writer, "invalid lat", http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(q.Get("lng"), 64)
	if err != nil {
		http.Error(writer, "invalid lng", http.StatusBadRequest)
		return
	}

	radius, err := strconv.Atoi(q.Get("radius_meters"))
	if err != nil || radius <= 0 {
		http.Error(writer, "invalid radius_meters", http.StatusBadRequest)
		return
	}

	limit := 20
	if q.Get("limit") != "" {
		if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 {
			limit = l
		}
	}

	types := q["types"] // supports ?types=article&types=event

	cursorStr := q.Get("cursor")

	input, err := toGetFeedInput(lat, lng, radius, types, limit, cursorStr)
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

	response, err := toGetFeedResponse(feedData)
	if err != nil {
		slog.Error("Failed to convert feed data to response", "error", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

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

// decodePointWKB decodes a POINT geography WKB into lat/lng.
func decodePointWKB(wkbBytes []byte) (lat float64, lng float64, err error) {
	g, err := wkb.Unmarshal(wkbBytes)
	if err != nil {
		return 0, 0, err
	}

	point, ok := g.(*geom.Point)
	if !ok {
		return 0, 0, errors.New("geometry is not a point")
	}

	coords := point.Coords()
	return coords[1], coords[0], nil // lat, lng
}

func toGetFeedResponse(feedData *entities.Feed) (GetFeedResponse, error) {
	items := make([]FeedItemDTO, 0, len(feedData.Items))

	for _, it := range feedData.Items {
		lat, lng, err := decodePointWKB(it.LocationWKB)
		if err != nil {
			return GetFeedResponse{}, err
		}

		items = append(items, FeedItemDTO{
			ID:          hex.EncodeToString(it.ID), // UUID v7 → hex string
			Type:        string(it.Type),
			Lat:         lat,
			Lng:         lng,
			PublishedAt: it.PublishedAt,
			Attributes:  it.Attributes,
			Score:       it.Score,
		})
	}

	var nextEncodedCursor string
	var err error
	if feedData.NextCursor != nil {
		nextEncodedCursor, err = entities.EncodeCursor(feedData.NextCursor)
		if err != nil {
			slog.Error("Failed to encode next cursor", "error", err)
			return GetFeedResponse{}, err
		}
	}

	return GetFeedResponse{
		Items:      items,
		NextCursor: nextEncodedCursor,
	}, nil
}

func toGetFeedInput(lat float64, lng float64, radiusMeters int, types []string, limit int, cursorStr string) (feed.GetFeedInput, error) {
	var cursor *entities.Cursor
	if cursorStr != "" {
		c, err := entities.DecodeCursor(cursorStr)
		if err != nil {
			return feed.GetFeedInput{}, err
		}
		cursor = c
	}

	inputTypes := make([][]byte, 0, len(types))

	for _, t := range types {
		inputTypes = append(inputTypes, []byte(t))
	}

	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return feed.GetFeedInput{
		Lat:          lat,
		Lng:          lng,
		RadiusMeters: radiusMeters,
		Types:        inputTypes,
		Limit:        limit,
		Cursor:       cursor,
	}, nil
}
