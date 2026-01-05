package entities

import "time"

type IngestionContent struct {
	ID []byte // UUID (16 bytes, raw)

	Type []byte // content type (e.g. "article", "event")

	// Location is stored as WKB (Well-Known Binary) returned by PostGIS
	// Example: ST_AsBinary(location)
	LocationWKB []byte

	PublishedAt time.Time
	CreatedAt   time.Time

	// Arbitrary metadata (JSONB), kept as raw bytes
	Attributes []byte

	// Domain-defined base relevance
	BaseScore float64
}
