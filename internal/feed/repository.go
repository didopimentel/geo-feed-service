package feed

import (
	"context"
	"fmt"
	"geo-feed-service/internal/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IRepository interface {
	GetFeed(ctx context.Context, q FeedQuery) (*entities.Feed, error)
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

type FeedQuery struct {
	Lat          float64
	Lng          float64
	RadiusMeters int
	Types        [][]byte
	Limit        int
	Cursor       *entities.Cursor
}

func (r *Repository) GetFeed(
	ctx context.Context,
	q FeedQuery,
) (*entities.Feed, error) {

	limit := q.Limit + 1 // fetch one extra row to detect next page

	sql := `
	WITH ranked AS (
		SELECT
			id,
			external_id,
			type,
			ST_AsBinary(location) AS location,
			published_at,
			created_at,
			attributes,
			base_score,

			-- distance in meters
			ST_Distance(
				location,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
			) AS distance_m
		FROM geo_content
		WHERE
			ST_DWithin(
				location,
				ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
				$3
			)
			AND type = ANY($4)
	),
	scored AS (
		SELECT
			*,
			(
				base_score
				* (1 / (1 + distance_m / 1000))
				* exp(-EXTRACT(EPOCH FROM (now() - published_at)) / 3600)
			) AS score
		FROM ranked
	)
	SELECT
		id,
		external_id,
		type,
		location,
		published_at,
		created_at,
		attributes,
		base_score,
		score
	FROM scored
	`

	args := []any{
		q.Lng,
		q.Lat,
		q.RadiusMeters,
		q.Types,
	}

	// Keyset pagination
	if q.Cursor != nil {
		sql += `
		WHERE (
			score < $5
			OR (
				score = $5
				AND published_at < $6
			)
			OR (
				score = $5
				AND published_at = $6
				AND id < $7
			)
		)
		`

		args = append(
			args,
			q.Cursor.Score,
			q.Cursor.PublishedAt,
			q.Cursor.ID,
		)
	}

	sql += `
	ORDER BY
		score DESC,
		published_at DESC,
		id DESC
	LIMIT $` + fmt.Sprint(len(args)+1)

	args = append(args, limit)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]entities.FeedItem, 0, limit)

	for rows.Next() {
		var e entities.FeedItem

		if err := rows.Scan(
			&e.ID,
			&e.ExternalID,
			&e.Type,
			&e.LocationWKB,
			&e.PublishedAt,
			&e.CreatedAt,
			&e.Attributes,
			&e.BaseScore,
			&e.Score,
		); err != nil {
			return nil, err
		}

		items = append(items, e)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var nextCursor *entities.Cursor

	if len(items) == limit {
		last := items[q.Limit]

		nextCursor = &entities.Cursor{
			Score:       last.Score,
			PublishedAt: last.PublishedAt,
			ID:          last.ID,
		}

		items = items[:q.Limit]
	}

	return &entities.Feed{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}
