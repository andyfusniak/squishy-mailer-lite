package store

import (
	"context"
	"database/sql/driver"
	"errors"
	"time"
)

type Repository interface {
	ProjectsRepository
	SMTPTransportsRepository
	GroupsRepository
	TemplatesRepository
}

//
// projects
//

var (
	// ErrProjectNotFound is returned when a project is not found.
	ErrProjectNotFound = errors.New("project not found")
)

// ProjectsRepository is the interface for the projects repository.
type ProjectsRepository interface {
	// InsertProject inserts a new project into the store
	InsertProject(ctx context.Context, params AddProject) (*Project, error)
}

// Project represents an individual project.
type Project struct {
	ProjectID   string
	ProjectName string
	Description string
	CreatedAt   Datetime
}

// AddProject is the input parameters for the InsertProject method.
type AddProject struct {
	ProjectID   string
	ProjectName string
	Description string
	CreatedAt   Datetime
}

const RFC3339Micro = "2006-01-02T15:04:05.000000Z07:00" // .000000Z = keep trailing zeros

// Datetime is a custom type for time.Time that can be scanned from the database.
type Datetime time.Time

func (t *Datetime) Scan(v any) error {
	vt, err := time.Parse(RFC3339Micro, v.(string))
	if err != nil {
		return err
	}
	*t = Datetime(vt)
	return nil
}

func (t *Datetime) Value() (driver.Value, error) {
	return time.Time(*t).UTC().Format(RFC3339Micro), nil
}

//
// smtp transports
//

type SMTPTransportsRepository interface {
	// InsertSMTPTransport inserts a new SMTP transport into the store.
	InsertSMTPTransport(ctx context.Context, params AddSMTPTransport) (*SMTPTransport, error)
}

// SMTPTransport represents an SMTP transport for a project.
type SMTPTransport struct {
	SMTPTransportID   string
	ProjectID         string
	TransportName     string
	Host              string
	Port              int
	Username          string
	EncryptedPassword string
	EmailFrom         string
	EmailReplyTo      string
	CreatedAt         Datetime
	ModifiedAt        Datetime
}

// AddSMTPTransport is the input parameters for the InsertSMTPTransport method.
type AddSMTPTransport struct {
	SMTPTransportID   string
	ProjectID         string
	TransportName     string
	Host              string
	Port              int
	Username          string
	EncryptedPassword string
	EmailFrom         string
	EmailReplyTo      string
	CreatedAt         Datetime
	ModifiedAt        Datetime
}

//
// groups
//

var (
	// ErrGroupNotFound is returned when a group is not found.
	ErrGroupNotFound = errors.New("group not found")
)

type GroupsRepository interface {
	// InsertGroup inserts a new group into the store
	InsertGroup(ctx context.Context, params AddGroup) (*Group, error)
}

// Group represents a group of templates.
type Group struct {
	GroupID    string
	ProjectID  string
	GroupName  string
	CreatedAt  Datetime
	ModifiedAt Datetime
}

// AddGroup logically groups together a set of email templates.
type AddGroup struct {
	GroupID    string
	ProjectID  string
	GroupName  string
	CreatedAt  Datetime
	ModifiedAt Datetime
}

//
// templates
//

var (
	// ErrTemplateNotFound is returned when a template is not found.
	ErrTemplateNotFound = errors.New("template not found")
)

type TemplatesRepository interface {
	// InsertTemplate inserts a new template into the store
	InsertTemplate(ctx context.Context, params AddTemplate) (*Template, error)

	// GetTemplate gets a template from the store.
	GetTemplate(ctx context.Context, projectID, templateID string) (*Template, error)
}

// Template represents an email template based on the schema.
type Template struct {
	TemplateID string
	GroupID    string
	ProjectID  string
	Txt        string
	HTML       string
	CreatedAt  Datetime
	ModifiedAt Datetime
}

// AddTemplate is the input parameters for the InsertTemplate method.
type AddTemplate struct {
	TemplateID string
	GroupID    string
	ProjectID  string
	Txt        string
	HTML       string
	CreatedAt  Datetime
	ModifiedAt Datetime
}
