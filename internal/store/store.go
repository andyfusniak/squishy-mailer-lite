package store

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Repository interface {
	ProjectsRepository
	SMTPTransportsRepository
	GroupsRepository
	TemplatesRepository
	Close() error
}

//
// projects
//

// create a list of error codes
const (
	ErrProjectAlreadyExists = "project_already_exists"
	ErrProjectNotFound      = "project_not_found"
	ErrGroupNotFound        = "group_not_found"
)

// ErrCode is a custom type for error codes.
type ErrCode string

var mapErrCodeToMessage = map[ErrCode]string{
	ErrProjectAlreadyExists: "project already exists",
	ErrProjectNotFound:      "project not found",
	ErrGroupNotFound:        "group not found",
}

// ServiceError is a custom error type.
type Error struct {
	Code ErrCode
	Msg  string
	err  error
}

// Error returns the error message.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s\n", e.Code, mapErrCodeToMessage[e.Code])
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.err
}

// NewStoreError creates a new Error with a code and an error.
func NewStoreError(code ErrCode, err error) *Error {
	return &Error{
		Code: code,
		Msg:  mapErrCodeToMessage[code],
		err:  err,
	}
}

// ProjectsRepository is the interface for the projects repository.
type ProjectsRepository interface {
	// InsertProject inserts a new project into the store
	InsertProject(ctx context.Context, params AddProject) (*Project, error)

	// GetProject gets a project from the store.
	GetProject(ctx context.Context, projectID string) (*Project, error)
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

// Scan parses a time from the database into a Datetime.
func (t *Datetime) Scan(v any) error {
	vt, err := time.Parse(RFC3339Micro, v.(string))
	if err != nil {
		return err
	}
	*t = Datetime(vt)
	return nil
}

// Value returns the time in the format expected by the database.
func (t *Datetime) Value() (driver.Value, error) {
	return time.Time(*t).UTC().Format(RFC3339Micro), nil
}

type JSONArray []string

// Scan unmarshals a JSON array into a JSONArray.
func (a *JSONArray) Scan(v any) error {
	// unmarshal the JSON array
	var arr []string
	if err := json.Unmarshal([]byte(v.(string)), &arr); err != nil {
		return err
	}
	*a = arr
	return nil
}

// Value returns the JSON array as a string.
func (a JSONArray) Value() (driver.Value, error) {
	v, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

//
// smtp transports
//

var (
	// ErrTransportNotFound is returned when an SMTP transport is not found.
	ErrTransportNotFound = errors.New("transport not found")
)

type SMTPTransportsRepository interface {
	// InsertSMTPTransport inserts a new SMTP transport into the store.
	InsertSMTPTransport(ctx context.Context, params AddSMTPTransport) (*SMTPTransport, error)
	GetSMTPTransport(ctx context.Context, transportID, projectID string) (*SMTPTransport, error)
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
	EmailFromName     string
	EmailReplyTo      JSONArray
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
	EmailFromName     string
	EmailReplyTo      JSONArray
	CreatedAt         Datetime
	ModifiedAt        Datetime
}

//
// groups
//

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
	TxtDigest  string
	HTML       string
	HTMLDigest string
	CreatedAt  Datetime
	ModifiedAt Datetime
}

// AddTemplate is the input parameters for the InsertTemplate method.
type AddTemplate struct {
	TemplateID string
	GroupID    string
	ProjectID  string
	Txt        string
	TxtDigest  string
	HTML       string
	HTMLDigest string
	CreatedAt  Datetime
	ModifiedAt Datetime
}
