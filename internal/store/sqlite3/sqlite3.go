package sqlite3

import (
	"context"
	"database/sql"
	"time"

	"github.com/andyfusniak/squishy-mailer-lite/internal/store"
	"github.com/pkg/errors"
)

// Store memberships store.
type Store struct {
	*Queries
	rw *sql.DB
}

// NewStore returns a new store.
func NewStore(ro, rw *sql.DB) *Store {
	return &Store{
		rw:      rw,
		Queries: NewQueries(ro, rw),
	}
}

// InsertProject inserts a new project into the store.
func (q *Queries) InsertProject(ctx context.Context, params store.AddProject) (*store.Project, error) {
	const query = `
INSERT INTO projects
  (id, pname, description, created_at)
VALUES
  (:id, :pname, :description, :created_at)
RETURNING
  id, pname, description, created_at
`
	var r store.Project
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("id", params.ID),
		sql.Named("pname", params.PName),
		sql.Named("description", params.Description),
		sql.Named("created_at", &now),
	).Scan(
		&r.ID,
		&r.PName,
		&r.Description,
		&r.CreatedAt,
	); err != nil {
		return nil, errors.Wrapf(err,
			"[sqlite3:projects] query row scan failed query=%q", query)
	}
	return &r, nil
}
