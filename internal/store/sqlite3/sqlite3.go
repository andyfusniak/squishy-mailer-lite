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
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// Store provides all functions to execute database queries and transactions.
type Store struct {
	*Queries
	readwrite *sql.DB
}

// NewStore returns a new store.
func NewStore(ro, rw *sql.DB) *Store {
	return &Store{
		Queries:   NewQueries(ro, rw),
		readwrite: rw,
	}
}

func (s *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.readwrite.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
	}
	q := s.withTx(tx)
	if err = fn(q); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("[postgres] tx rollback failed: %v: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

// Close the store.
func (q *Queries) Close() error {
	var isReadOnlyErr, isReadWriteErr bool

	// convert the interface to its underlying type and check for errors
	rw := q.readwrite.(*sql.DB)
	if err := rw.Close(); err != nil {
		isReadWriteErr = true
	}

	ro := q.readonly.(*sql.DB)
	if err := ro.Close(); err != nil {
		isReadWriteErr = true
	}

	// report any errors
	if isReadOnlyErr || isReadWriteErr {
		if isReadOnlyErr && isReadWriteErr {
			return errors.New("failed to close both database connections")
		} else if isReadWriteErr {
			return errors.New("failed to close the read-write database connection")
		} else if isReadOnlyErr {
			return errors.New("failed to close the read-only database connection")
		}
	}

	return nil
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

//
// projects
//

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
		if serr, ok := err.(sqlite3.Error); ok {
			if serr.Code == sqlite3.ErrConstraint &&
				serr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return nil, store.NewStoreError(store.ErrProjectAlreadyExists, err)
			}
		}
		return nil, errors.Wrapf(err,
			"[sqlite3:projects] query row scan failed query=%q", query)
	}
	return &r, nil
}

// GetProject gets a project from the store by projectID. If the project is
// not found, an error of type store.ErrProjectNotFound is returned.
func (q *Queries) GetProject(ctx context.Context, projectID string) (*store.Project, error) {
	const query = `
select
  p.project_id, p.project_name, description, p.created_at
from projects as p
where
  p.project_id = :project_id
`
	var r store.Project
	if err := q.readonly.QueryRowContext(ctx, query,
		sql.Named("project_id", projectID),
	).Scan(
		&r.ProjectID,
		&r.ProjectName,
		&r.Description,
		&r.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.NewStoreError(store.ErrProjectNotFound, err)
		}

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
insert into smtp_transports as t (
  smtp_transport_id, project_id, transport_name, host, port, username,
  encrypted_password, email_from, email_from_name, email_replyto,
  created_at, modified_at
)
select
  :smtp_transport_id as smtp_transport_id,
  p.project_id as project_id,
  :transport_name as transport_name,
  :host as host,
  :port as port,
  :username as username,
  :encrypted_password as encrypted_password,
  :email_from as email_from,
  :email_from_name as email_from_name,
  :email_replyto as email_replyto,
  :created_at as created_at,
  :modified_at as modified_at
from projects as p
where p.project_id = :project_id
returning
  smtp_transport_id, project_id, transport_name, host, port, username,
  encrypted_password, email_from, email_from_name, email_replyto,
  created_at, modified_at
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
		sql.Named("email_from_name", params.EmailFromName),
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
		&r.EmailFromName,
		&r.EmailReplyTo,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		return nil, errors.Wrapf(err,
			"[sqlite3:smtp_transports] query row scan failed query=%q", query)
	}
	return &r, nil
}

// GetSMTPTransport gets a SMTP transport from the store by composite primary
// key (transportID, projectID).
func (q *Queries) GetSMTPTransport(ctx context.Context, transportID, projectID string) (*store.SMTPTransport, error) {
	const query = `
select
  coalesce(t.smtp_transport_id, '') as smtp_transport_id,
  p.project_id,
  coalesce(t.transport_name, '') as transport_name,
  coalesce(t.host, '') as host,
  coalesce(t.port, 0) as port,
  coalesce(t.username, '') as username,
  coalesce(t.encrypted_password, '') as encrypted_password,
  coalesce(t.email_from, '') as email_from,
  coalesce(t.email_from_name, '') as email_from_name,
  coalesce(t.email_replyto, '') as email_replyto,
  coalesce(t.created_at, '1970-01-01T00:00:00.000000Z') as created_at,
  coalesce(t.modified_at, '1970-01-01T00:00:00.000000Z') as modified_at
from projects as p
left outer join smtp_transports as t
  on p.project_id = t.project_id and t.smtp_transport_id = :smtp_transport_id
where
  p.project_id = :project_id
`

	var r store.SMTPTransport
	if err := q.readonly.QueryRowContext(ctx, query,
		sql.Named("project_id", projectID),
		sql.Named("smtp_transport_id", transportID),
	).Scan(
		&r.SMTPTransportID,
		&r.ProjectID,
		&r.TransportName,
		&r.Host,
		&r.Port,
		&r.Username,
		&r.EncryptedPassword,
		&r.EmailFrom,
		&r.EmailFromName,
		&r.EmailReplyTo,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		// if there are no rows returned, then the project does not exist
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.NewStoreError(store.ErrProjectNotFound, err)
		}

		return nil, errors.Wrapf(err,
			"[sqlite3:smtp_transports] query row scan failed query=%q", query)
	}

	if r.SMTPTransportID == "" {
		return nil, store.ErrTransportNotFound
	}

	return &r, nil
}

//
// groups
//

// InsertGroup inserts a new group into the store.
func (q *Queries) InsertGroup(ctx context.Context, params store.AddGroup) (*store.Group, error) {
	const query = `
insert into groups
  (group_id, project_id, group_name, created_at, modified_at)
values
  (:group_id, :project_id, :group_name, :created_at, :modified_at)
returning
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
		// if sqlite3 returns a foreign key constraint error, then the project does not existing
		// assert the underlying sqlite3 type
		if serr, ok := err.(sqlite3.Error); ok {
			// In the C API for SQLite, it is not directly possible to determine
			// which specific foreign key constraint failed when multiple
			// constraints are violated. The error message that is returned by
			// SQLite does not provide this level of detail. However, since
			// there is only one foreign key constraint in this case, we can
			// assume that the constraint that failed was the foreign key
			// constraint that references the projects table.
			//
			// see https://www.sqlite.org/rescode.html#constraint_foreignkey
			if serr.Code == sqlite3.ErrConstraint && serr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
				return nil, store.NewStoreError(store.ErrProjectNotFound, serr)
			}
		}

		return nil, errors.Wrapf(err,
			"[sqlite3:groups] query row scan failed query=%q", query)
	}
	return &r, nil
}

// GetGroup gets a group from the store.
func (q *Queries) GetGroup(ctx context.Context, projectID, groupID string) (*store.Group, error) {
	const query = `
select
  coalesce(g.group_id, '') as group_id,
  p.project_id,
  coalesce(g.group_name, '') as group_name,
  coalesce(g.created_at, '1970-01-01T00:00:00.000000Z') as created_at,
  coalesce(g.modified_at, '1970-01-01T00:00:00.000000Z') as modified_at
from projects as p
left outer join groups as g
  on p.project_id = g.project_id
  and g.group_id = :group_id
where
  p.project_id = :project_id
`
	var r store.Group
	if err := q.readonly.QueryRowContext(ctx, query,
		sql.Named("project_id", projectID),
		sql.Named("group_id", groupID),
	).Scan(
		&r.GroupID,
		&r.ProjectID,
		&r.GroupName,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		// if there are no rows returned, then the project does not exist
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.NewStoreError(store.ErrProjectNotFound, err)
		}
		return nil, errors.Wrapf(err,
			"[sqlite3:groups] query row scan failed query=%q", query)
	}

	if r.GroupID == "" {
		return nil, store.NewStoreError(store.ErrGroupNotFound, nil)
	}

	return &r, nil
}

//
// templates
//

// InsertTemplate inserts a new template into the store.
func (q *Queries) InsertTemplate(ctx context.Context, params store.AddTemplate) (*store.Template, error) {
	const query = `
insert into templates
  (template_id, group_id, project_id, txt, txt_digest, html, html_digest, created_at, modified_at)
values
  (:template_id, :group_id, :project_id, :txt, :txt_digest, :html, :html_digest, :created_at, :modified_at)
returning
  template_id, group_id, project_id, txt, txt_digest, html, html_digest, created_at, modified_at
`
	var r store.Template
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("template_id", params.TemplateID),
		sql.Named("group_id", params.GroupID),
		sql.Named("project_id", params.ProjectID),
		sql.Named("txt", params.Txt),
		sql.Named("txt_digest", params.TxtDigest),
		sql.Named("html", params.HTML),
		sql.Named("html_digest", params.HTMLDigest),
		sql.Named("created_at", &now),
		sql.Named("modified_at", &now),
	).Scan(
		&r.TemplateID,
		&r.GroupID,
		&r.ProjectID,
		&r.Txt,
		&r.TxtDigest,
		&r.HTML,
		&r.HTMLDigest,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		return nil, errors.Wrapf(err,
			"[sqlite3:templates] query row scan failed query=%q", query)
	}
	return &r, nil
}

// SetTemplate sets a template in the store. If the template does not exist
// it will be created. If the template does exist and the digests are the same
// as the ones provided by the caller, then the template will not be updated.
// If the digests are different, then the template will be updated.
func (s *Store) SetTemplate(ctx context.Context, params store.SetTemplateParams) (*store.Template, error) {
	const chkDigestQuery = `
select
  coalesce(t.template_id, '') as template_id,
  coalesce(t.group_id, '') as group_id,
  p.project_id,
  coalesce(txt_digest == :txt_digest, FALSE) as txt_digest_eq,
  coalesce(html_digest == :html_digest, FALSE) as html_digest_eq,
  coalesce(t.created_at, '1970-01-01T00:00:00.000000Z') as created_at,
  coalesce(t.modified_at, '1970-01-01T00:00:00.000000Z') as modified_at
from projects as p
left outer join templates as t
  on p.project_id = t.project_id and t.template_id = :template_id
where
  p.project_id = :project_id
`
	var r *store.Template
	if err := s.execTx(ctx, func(q *Queries) error {
		// 1. get the txt and html digest for the template
		// if no rows are returned then the project does not exist
		// if one row is returned and the template id is empty
		// then the template does not exist
		// otherwise txt_digest_eq and html_digest_eq will indicate
		// if the digests are equal to the ones provided by the caller
		//
		// only use the q.readwrite connection for this query
		// because the readonly query will not see the uncommitted
		// changes made by the insert query
		var templateID, groupID, projectID string
		var txtDigestEq, htmlDigestEq bool
		var createdAt, modifiedAt store.Datetime
		if err := q.readwrite.QueryRowContext(ctx, chkDigestQuery,
			sql.Named("txt_digest", params.TxtDigest),
			sql.Named("html_digest", params.HTMLDigest),
			sql.Named("project_id", params.ProjectID),
			sql.Named("template_id", params.TemplateID),
		).Scan(
			&templateID,
			&groupID,
			&projectID,
			&txtDigestEq,
			&htmlDigestEq,
			&createdAt,
			&modifiedAt,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return store.NewStoreError(store.ErrProjectNotFound, err)
			}
			return errors.Wrapf(err,
				"[sqlite3:templates] query row scan failed query=%q", chkDigestQuery)
		}

		if templateID == "" {
			// the template does not exist
			// 2. create a new template
			var err error
			r, err = q.InsertTemplate(ctx, store.AddTemplate{
				TemplateID: params.TemplateID,
				GroupID:    params.GroupID,
				ProjectID:  params.ProjectID,
				Txt:        params.Txt,
				TxtDigest:  params.TxtDigest,
				HTML:       params.HTML,
				HTMLDigest: params.HTMLDigest,
				CreatedAt:  store.Datetime(time.Now().UTC()),
				ModifiedAt: store.Datetime(time.Now().UTC()),
			})
			if err != nil {
				return err
			}

			return nil
		}

		// 2. the template exists and the digests are the same so there is no
		// need to update the template (or 3 below)
		if txtDigestEq && htmlDigestEq {
			r = &store.Template{
				TemplateID: params.TemplateID,
				GroupID:    groupID,
				ProjectID:  params.ProjectID,
				Txt:        params.Txt,
				TxtDigest:  params.TxtDigest,
				HTML:       params.HTML,
				HTMLDigest: params.HTMLDigest,
				CreatedAt:  createdAt,
				ModifiedAt: modifiedAt,
			}
			return nil
		}

		// 3. the digests differ so update the template
		var err error
		r, err = q.updateTemplate(ctx, updateTemplateParams{
			projectID:  params.ProjectID,
			templateID: params.TemplateID,
			txt:        params.Txt,
			txtDigest:  params.TxtDigest,
			html:       params.HTML,
			htmlDigest: params.HTMLDigest,
		})
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return r, nil
}

type updateTemplateParams struct {
	projectID  string
	templateID string
	txt        string
	txtDigest  string
	html       string
	htmlDigest string
}

func (q *Queries) updateTemplate(ctx context.Context, params updateTemplateParams) (*store.Template, error) {
	const query = `
update templates
set
  txt = :txt, txt_digest = :txt_digest,
  html = :html, html_digest = :html_digest,
  modified_at = :modified_at
where
  template_id = :template_id and project_id = :project_id
returning
  template_id, group_id, project_id, txt, txt_digest, html, html_digest, created_at, modified_at
`
	var r store.Template
	now := store.Datetime(time.Now().UTC())
	if err := q.readwrite.QueryRowContext(ctx, query,
		sql.Named("txt", params.txt),
		sql.Named("txt_digest", params.txtDigest),
		sql.Named("html", params.html),
		sql.Named("html_digest", params.htmlDigest),
		sql.Named("modified_at", &now),
		sql.Named("template_id", params.templateID),
		sql.Named("project_id", params.projectID),
	).Scan(
		&r.TemplateID,
		&r.GroupID,
		&r.ProjectID,
		&r.Txt,
		&r.TxtDigest,
		&r.HTML,
		&r.HTMLDigest,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		return nil, errors.Wrapf(err,
			"[sqlite3:templates] query row scan failed query=%q", query)
	}
	return &r, nil
}

// GetTemplate gets a template from the store by projectID and templateID.
// Templates are unique within a project. If the project is not found, an
// error of type store.ErrProjectNotFound is returned. If the template is
// not found, the error will be of type store.ErrTemplateNotFound.
func (q *Queries) GetTemplate(ctx context.Context, projectID, templateID string) (*store.Template, error) {
	const query = `
select
  coalesce(t.template_id, '') as template_id,
  coalesce(t.group_id, '') as group_id,
  p.project_id,
  coalesce(t.txt, '') as txt,
  coalesce(t.html, '') as html,
  coalesce(t.created_at, '1970-01-01T00:00:00.000000Z') as created_at,
  coalesce(t.modified_at, '1970-01-01T00:00:00.000000Z') as modified_at
from projects as p
left outer join templates as t
  on p.project_id = t.project_id and t.template_id = :template_id
where
  p.project_id = :project_id
`
	var r store.Template
	if err := q.readonly.QueryRowContext(ctx, query,
		sql.Named("project_id", projectID),
		sql.Named("template_id", templateID),
	).Scan(
		&r.TemplateID,
		&r.GroupID,
		&r.ProjectID,
		&r.Txt,
		&r.HTML,
		&r.CreatedAt,
		&r.ModifiedAt,
	); err != nil {
		// if there are no rows returned, then the project does not exist
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.NewStoreError(store.ErrProjectNotFound, err)
		}

		return nil, errors.Wrapf(err,
			"[sqlite3:templates] query row scan failed query=%q", query)
	}

	if r.TemplateID == "" {
		return nil, store.NewStoreError(store.ErrTemplateNotFound, nil)
	}

	return &r, nil
}
