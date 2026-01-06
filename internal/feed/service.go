package feed

import (
	"context"
	"geo-feed-service/internal/entities"
)

type Service struct {
	repository IRepository
}

func NewService(repo IRepository) *Service {
	return &Service{repository: repo}
}

type GetFeedInput struct {
	// Optional location filter (latitude, longitude, radius in meters)
	Lat          float64
	Lng          float64
	RadiusMeters int
	// Optional content types filter (array of domain type bytea)
	Types [][]byte
	// Maximum number of items to return
	Limit int
	// Optional pagination cursor
	Cursor *entities.Cursor
}

func (s *Service) GetFeed(ctx context.Context, input GetFeedInput) (*entities.Feed, error) {
	return s.repository.GetFeed(
		ctx,
		FeedQuery{
			Lat:          input.Lat,
			Lng:          input.Lng,
			RadiusMeters: input.RadiusMeters,
			Types:        input.Types,
			Limit:        input.Limit,
			Cursor:       input.Cursor,
		},
	)
}
