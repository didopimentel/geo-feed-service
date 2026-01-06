package ingestion

import "time"

type Service struct {
	repository IRepository
}

func NewService(repo IRepository) *Service {
	return &Service{repository: repo}
}

type IngestDataInput struct {
	// Optional external ID (publisher ID, event ID, etc.)
	ExternalID []byte

	// Domain type: article, event, alert, etc.
	Type []byte

	// Location as raw coordinates (not WKB yet)
	Lat float64
	Lng float64

	// When this item should become visible to users
	PublishedAt time.Time

	// Arbitrary metadata (JSON)
	Attributes []byte

	// Optional base relevance (defaults to 1.0)
	BaseScore float64
}

func (s *Service) IngestData(input IngestDataInput) error {
	return s.repository.SaveContent(input.ExternalID, input.Type, input.Lat, input.Lng, input.PublishedAt, input.Attributes, input.BaseScore)
}
