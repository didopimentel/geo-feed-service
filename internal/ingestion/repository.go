package ingestion

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// SaveContent inserts a new geo content into Postgres.
//
// Spatial encoding (lat/lng → geography) is handled here.
// created_at and id are database-managed.
func (r *Repository) SaveContent(
	externalID []byte,
	domainType []byte,
	lat float64,
	lng float64,
	publishedAt time.Time,
	attributes []byte,
	baseScore float64,
) error {
	const query = `
		INSERT INTO feed_items (
			external_id,
			type,
			location,
			published_at,
			attributes,
			base_score
		)
		VALUES (
			$1,
			$2,
			ST_SetSRID(ST_MakePoint($3, $4), 4326)::geography,
			$5,
			$6,
			$7
		)
		ON CONFLICT (external_id) DO NOTHING
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		externalID,
		domainType,
		lng,
		lat,
		publishedAt,
		attributes,
		baseScore,
	)
	if err != nil {
		return err
	}

	// If row already exists, this will be 0 — and that's OK
	if cmdTag.RowsAffected() > 1 {
		return errors.New("unexpected number of rows affected")
	}

	return nil
}
