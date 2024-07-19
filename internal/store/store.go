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
	MailQueueRepository
	Close() error
}

//
// projects
//

// create a list of error codes
const (
	ErrProjectAlreadyExists  = "project_already_exists"
	ErrProjectNotFound       = "project_not_found"
	ErrSMTPTransportNotFound = "smtp_transport_not_found"
	ErrGroupNotFound         = "group_not_found"
	ErrTemplateNotFound      = "template_not_found"
)

// ErrCode is a custom type for error codes.
type ErrCode string

var mapErrCodeToMessage = map[ErrCode]string{
	ErrProjectAlreadyExists:  "project already exists",
	ErrProjectNotFound:       "project not found",
	ErrSMTPTransportNotFound: "smtp transport not found",
	ErrGroupNotFound:         "group not found",
	ErrTemplateNotFound:      "template not found",
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
type Datetime struct {
	time.Time
}

// Scan parses a time from the database into a Datetime.
func (t *Datetime) Scan(v any) error {
	vt, err := time.Parse(RFC3339Micro, v.(string))
	if err != nil {
		return err
	}
	*t = Datetime{vt}
	return nil
}

// Value returns the time in the format expected by the database.
func (t *Datetime) Value() (driver.Value, error) {
	return (*t).UTC().Format(RFC3339Micro), nil
}

// JSONUnmarshal unmarshals a JSON string into a Datetime.
func (t *Datetime) JSONUnmarshal(data []byte) error {
	fmt.Printf("*******************************")
	fmt.Printf("data: %s\n", string(data))
	*t = Datetime{time.Now()}
	// vt, err := time.Parse(RFC3339Micro, string(data))
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("%#v\n", vt)
	// *t = Datetime(vt)
	return nil
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

type TemplatesRepository interface {
	// InsertTemplate inserts a new template into the store
	InsertTemplate(ctx context.Context, params AddTemplate) (*Template, error)

	// SetTemplate sets a template in the store. If the template does not exist, it is created.
	// If the template exists, it is updated if the digests do not match.
	SetTemplate(ctx context.Context, params SetTemplateParams) (*Template, error)

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

// SetTemplateParams is the input parameters for the SetTemplateParams method.
type SetTemplateParams struct {
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

// TemplateDigest is a digest of a template.
type TemplateDigest struct {
	TemplateID string
	TxtDigest  string
	HTMLDigest string
}

// mail queue

const (
	// MailQueueStateQueued represents the state of an email in the mail queue when it is queued.
	MailStateQueued = "queued"
)

type MailQueueRepository interface {
	// InsertMailQueue inserts a new email into the mail queue.
	InsertMailQueue(ctx context.Context, params AddMailQueue) (*MailQueue, error)
}

// MailQueue represents an email in the mail queue.
type MailQueue struct {
	MailQueueID string
	ProjectID   string
	Mstate      string
	Subj        string
	EmailTo     JSONArray
	Body        string
	Transport   MailQueueSMTPTransport
	Metadata    MailQueueMetadata
	CreatedAt   Datetime
	ModifiedAt  Datetime
}

// AddMailQueue is the input parameters for the InsertMailQueue method.
type AddMailQueue struct {
	MailQueueID     string
	ProjectID       string
	Body            string
	SMTPTransportID string
	TemplateID      string
	Subj            string
	EmailTo         JSONArray
}

// MailQueueBody represents the body of an email in the mail queue.
type MailQueueBody struct {
	Txt            string            `json:"txt"`
	TxtDigest      string            `json:"txt_digest"`
	HTML           string            `json:"html"`
	HTMLDigest     string            `json:"html_digest"`
	TemplateParams map[string]string `json:"template_params"`
}

func (b *MailQueueBody) Scan(v any) error {
	var obj MailQueueBody
	if err := json.Unmarshal([]byte(v.(string)), &v); err != nil {
		return err
	}
	*b = obj
	return nil
}

// MailQueueMetadata represents the metadata of an email in the mail queue.
type MailQueueMetadata struct {
	Project  *MailQueueProject  `json:"project"`
	Group    *MailQueueGroup    `json:"group"`
	Template *MailQueueTemplate `json:"template"`
	// SMTPTransport MailQueueSMTPTransport `json:"smtp_transport"`
}

func (m *MailQueueMetadata) Scan(v any) error {
	obj := MailQueueMetadata{
		Project: &MailQueueProject{},
	}
	s := []byte(v.(string))
	fmt.Printf("%#v\n", string(s))
	if err := json.Unmarshal(s, &obj); err != nil {
		return err
	}
	*m = obj
	return nil
}

// MailQueueProject represents the project of an email at the time it was queued.
type MailQueueProject struct {
	ProjectID   string   `json:"project_id"`
	ProjectName string   `json:"project_name"`
	CreatedAt   Datetime `json:"created_at"`
}

func (p *MailQueueProject) UnmarshalJSON(v []byte) error {
	type Alias MailQueueProject
	var obj Alias
	if err := json.Unmarshal(v, &obj); err != nil {
		return err
	}
	*p = MailQueueProject(obj)
	return nil
}

// MailQueueGroup represents the group of an email at the time it was queued.
type MailQueueGroup struct {
	GroupID    string   `json:"group_id"`
	ProjectID  string   `json:"project_id"`
	GroupName  string   `json:"group_name"`
	CreatedAt  Datetime `json:"created_at"`
	ModifiedAt Datetime `json:"modified_at"`
}

// UnmarshalJSON unmarshals a JSON object into a MailQueueGroup.
func (g *MailQueueGroup) UnmarshalJSON(v []byte) error {
	type Alias MailQueueGroup
	var obj Alias
	if err := json.Unmarshal(v, &obj); err != nil {
		return err
	}
	*g = MailQueueGroup(obj)
	return nil
}

// MailQueueTemplate represents the template of an email at the time it was queued.
type MailQueueTemplate struct {
	TemplateID string   `json:"template_id"`
	GroupID    string   `json:"group_id"`
	ProjectID  string   `json:"project_id"`
	Txt        string   `json:"txt"`
	TxtDigest  string   `json:"txt_digest"`
	HTML       string   `json:"html"`
	HTMLDigest string   `json:"html_digest"`
	CreatedAt  Datetime `json:"created_at"`
	ModifiedAt Datetime `json:"modified_at"`
}

func (t *MailQueueTemplate) UnmarshalJSON(v []byte) error {
	type Alias MailQueueTemplate
	var obj Alias
	if err := json.Unmarshal(v, &obj); err != nil {
		return err
	}
	*t = MailQueueTemplate(obj)
	return nil
}

// MailQueueSMTPTransport represents the transport of an email at the time it was queued.
type MailQueueSMTPTransport struct {
	SMTPTransportID   string    `json:"smtp_transport_id"`
	ProjectID         string    `json:"project_id"`
	TransportName     string    `json:"transport_name"`
	Host              string    `json:"host"`
	Port              int       `json:"port"`
	Username          string    `json:"username"`
	EncryptedPassword string    `json:"encrypted_password"`
	EmailFrom         string    `json:"email_from"`
	EmailFromName     string    `json:"email_from_name"`
	EmailReplyTo      JSONArray `json:"email_reply_to"`
	CreatedAt         Datetime  `json:"created_at"`
	ModifiedAt        Datetime  `json:"modified_at"`
}

// Scan unmarshals a JSON object into a MailQueueSMTPTransport.
func (s *MailQueueSMTPTransport) Scan(v any) error {
	var obj MailQueueSMTPTransport
	if err := json.Unmarshal([]byte(v.(string)), &obj); err != nil {
		return err
	}
	*s = obj
	return nil
}
