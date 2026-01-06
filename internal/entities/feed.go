package entities

import "time"

// FeedItem represents a persisted feed item returned by feed queries.
type FeedItem struct {
	// Primary key (UUID v7, time-ordered)
	ID []byte // 16 bytes

	// Idempotency / external reference
	ExternalID []byte

	// Domain type (article, event, alert, etc.)
	Type []byte

	// Spatial data returned as WKB (PostGIS)
	// Usually produced via ST_AsBinary(location)
	LocationWKB []byte

	// Domain visibility time
	PublishedAt time.Time

	// System ingestion time
	CreatedAt time.Time

	// Arbitrary metadata (JSONB)
	Attributes []byte

	// Base relevance score
	BaseScore float64

	// Computed relevance score
	Score float64
}

type Cursor struct {
	Score       float64
	PublishedAt time.Time
	ID          []byte
}

type Feed struct {
	Items      []FeedItem
	NextCursor *Cursor
}
