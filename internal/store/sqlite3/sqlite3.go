package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/andyfusniak/squishy-mailer-lite/internal/store"
	"github.com/andyfusniak/squishy-mailer-lite/internal/store/sqlite3/schema"
	"github.com/golang-migrate/migrate/v4"
	driversqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
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

// CreateSQLiteDBSchema creates the tables using the schema for
// the sqlite3 database. If the tables already exist, this function
// will not modify them.
func CreateSqliteDBSchema(db *sql.DB) error {
	driver, err := driversqlite3.WithInstance(db, &driversqlite3.Config{NoTxWrap: true})
	if err != nil {
		return fmt.Errorf("failed to get new sqlite3 driver instance: %w", err)
	}

	source, err := httpfs.New(http.FS(schema.Migrations), "migrations")
	if err != nil {
		return err
	}

	mg, err := migrate.NewWithInstance("https", source, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to get new migrate instance: %w", err)
	}

	if err := mg.Up(); err != nil {
		return fmt.Errorf("migrate up failed: %w", err)
	}

	return nil
}

// InsertProject inserts a new project into the store.
func (q *Queries) InsertProject(ctx context.Context, params store.AddProject) (*store.Project, error) {
	const query = `
insert into projects
  (project_id, project_name, description, created_at)
values
  (:project_id, :project_name, :description, :created_at)
returning
  project_id, project_name, description, created_at
`
	var r store.Project
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("project_id", params.ProjectID),
		sql.Named("project_name", params.ProjectName),
		sql.Named("description", params.Description),
		sql.Named("created_at", &now),
	).Scan(
		&r.ProjectID,
		&r.ProjectName,
		&r.Description,
		&r.CreatedAt,
	); err != nil {
		return nil, errors.Wrapf(err,
			"[sqlite3:projects] query row scan failed query=%q", query)
	}
	return &r, nil
}

//
// smtp transports
//

// InsertSMTPTransport inserts a new SMTP transport into the store.
func (q *Queries) InsertSMTPTransport(ctx context.Context, params store.AddSMTPTransport) (*store.SMTPTransport, error) {
	const query = `
insert into smtp_transports as t (smtp_transport_id, project_id, transport_name, host, port, username, encrypted_password, email_from, email_replyto, created_at, modified_at)
select
  :smtp_transport_id as smtp_transport_id,
  p.project_id as project_id,
  :transport_name as transport_name,
  :host as host,
  :port as port,
  :username as username,
  :encrypted_password as encrypted_password,
  :email_from as email_from,
  :email_replyto as email_replyto,
  :created_at as created_at,
  :modified_at as modified_at
from projects as p
where p.project_id = :project_id
returning
  smtp_transport_id, project_id, transport_name, host, port, username, encrypted_password, email_from, email_replyto, created_at, modified_at
`
	var r store.SMTPTransport
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("smtp_transport_id", params.SMTPTransportID),
		sql.Named("transport_name", params.TransportName),
		sql.Named("host", params.Host),
		sql.Named("port", params.Port),
		sql.Named("username", params.Username),
		sql.Named("encrypted_password", params.EncryptedPassword),
		sql.Named("email_from", params.EmailFrom),
		sql.Named("email_replyto", params.EmailReplyTo),
		sql.Named("created_at", &now),
		sql.Named("modified_at", &now),
		sql.Named("project_id", params.ProjectID),
	).Scan(
		&r.SMTPTransportID,
		&r.ProjectID,
		&r.TransportName,
		&r.Host,
		&r.Port,
		&r.Username,
		&r.EncryptedPassword,
		&r.EmailFrom,
		&r.EmailReplyTo,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		return nil, errors.Wrapf(err, "[sqlite3:smtp_transports] query row scan failed query=%q", query)
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
  (group_id, project_id, group_name, created_at, modified_at)
VALUES
  (:group_id, :project_id, :group_name, :created_at, :modified_at)
RETURNING
  group_id, project_id, group_name, created_at, modified_at
	`
	var r store.Group
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("group_id", params.GroupID),
		sql.Named("project_id", params.ProjectID),
		sql.Named("group_name", params.GroupName),
		sql.Named("created_at", &now),
		sql.Named("modified_at", &now),
	).Scan(
		&r.GroupID,
		&r.ProjectID,
		&r.GroupName,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		return nil, errors.Wrapf(err, "[sqlite3:groups] query row scan failed query=%q", query)
	}
	return &r, nil
}

//
// templates
//

// InsertTemplate inserts a new template into the store.
func (q *Queries) InsertTemplate(ctx context.Context, params store.AddTemplate) (*store.Template, error) {
	const query = `
INSERT INTO templates
  (template_id, group_id, project_id, txt, html, created_at, modified_at)
VALUES
  (:template_id, :group_id, :project_id, :txt, :html, :created_at, :modified_at)
RETURNING
  template_id, group_id, project_id, txt, html, created_at, modified_at
`
	var r store.Template
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("template_id", params.TemplateID),
		sql.Named("group_id", params.GroupID),
		sql.Named("project_id", params.ProjectID),
		sql.Named("txt", params.Txt),
		sql.Named("html", params.HTML),
		sql.Named("created_at", &now),
		sql.Named("modified_at", &now),
	).Scan(
		&r.TemplateID,
		&r.GroupID,
		&r.ProjectID,
		&r.Txt,
		&r.HTML,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		return nil, errors.Wrapf(err, "[sqlite3:templates] query row scan failed query=%q", query)
	}
	return &r, nil
}
