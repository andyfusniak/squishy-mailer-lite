package store

import (
	"context"
	"database/sql/driver"
	"time"
)

type Repository interface {
	ProjectsRepository
	TransportsRepository
	// GroupsRepository
	// TemplatesRepository
}

//
// projects
//

// ProjectsRepository is the interface for the projects repository.
type ProjectsRepository interface {
	// InsertProject inserts a new project into the store
	InsertProject(ctx context.Context, params AddProject) (*Project, error)
}

// Project represents an individual project.
type Project struct {
	ID          string
	PName       string
	Description string
	CreatedAt   Datetime
}

// AddProject is the input parameters for the InsertProject method.
type AddProject struct {
	ID          string
	PName       string
	Description string
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
// transports
//

type TransportsRepository interface {
	// InsertTransport inserts a new transport into the store
	InsertTransport(ctx context.Context, params AddTransport) (*Transport, error)
}

// Transport represents an SMTP transport for a project.
type Transport struct {
	ID                string
	ProjectID         string
	TRName            string
	Host              string
	Port              int
	Username          string
	EncryptedPassword string
	EmailFrom         string
	EmailReplyTo      string
	CreatedAt         Datetime
	ModifiedAt        Datetime
}

// AddTransport is the input parameters for the InsertTransport method.
type AddTransport struct {
	ID                string
	ProjectID         string
	TRName            string
	Host              string
	Port              int
	Username          string
	EncryptedPassword string
	EmailFrom         string
	EmailReplyTo      string
}
