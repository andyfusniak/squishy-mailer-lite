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

//
// transports
//

// InsertTransport inserts a new transport into the store.
func (q *Queries) InsertTransport(ctx context.Context, params store.AddTransport) (*store.Transport, error) {
	const query = `
INSERT INTO transports
  (id, project_id, trname,host, port, username, encrypted_password, email_from, email_replyto, created_at, modified_at)
VALUES
  (:id, :project_id, :trname, :host, :port, :username, :encrypted_password, :email_from, :email_replyto, :created_at, :modified_at)
RETURNING
  id, project_id, trname, host, port, username, encrypted_password, email_from, email_replyto, created_at, modified_at
`
	var r store.Transport
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("id", params.ID),
		sql.Named("project_id", params.ProjectID),
		sql.Named("trname", params.TRName),
		sql.Named("host", params.Host),
		sql.Named("port", params.Port),
		sql.Named("username", params.Username),
		sql.Named("encrypted_password", params.EncryptedPassword),
		sql.Named("email_from", params.EmailFrom),
		sql.Named("email_replyto", params.EmailReplyTo),
		sql.Named("created_at", &now),
		sql.Named("modified_at", &now),
	).Scan(
		&r.ID,
		&r.ProjectID,
		&r.TRName,
		&r.Host,
		&r.Port,
		&r.Username,
		&r.EncryptedPassword,
		&r.EmailFrom,
		&r.EmailReplyTo,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		return nil, errors.Wrapf(err, "[sqlite3:transports] query row scan failed query=%q", query)
	}
	return &r, nil
}

//
// groups
//

// InsertGroup inserts a new group into the store.
func (q *Queries) InsertGroup(ctx context.Context, params store.AddGroup) (*store.Group, error) {
	const query = `
INSERT INTO groups
  (id, project_id, gname, created_at, modified_at)
VALUES
  (:id, :project_id, :gname, :created_at, :modified_at)
RETURNING
  id, project_id, gname, created_at, modified_at
	`
	var r store.Group
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("id", params.ID),
		sql.Named("project_id", params.ProjectID),
		sql.Named("gname", params.GName),
		sql.Named("created_at", &now),
		sql.Named("modified_at", &now),
	).Scan(
		&r.ID,
		&r.ProjectID,
		&r.GName,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		return nil, errors.Wrapf(err, "[sqlite3:groups] query row scan failed query=%q", query)
	}
	return &r, nil
}
