package entity

import (
	"fmt"
	"time"
)

const jsonTime = "2006-01-02T15:04:05.000Z07:00" // .000Z = keep trailing zeros

// ISOTime custom type to allow for JSON microsecond formating.
type ISOTime time.Time

// MarshalJSON provides microsecond formating
func (t ISOTime) MarshalJSON() ([]byte, error) {
	vt := time.Time(t)
	vt = vt.UTC().Round(time.Millisecond)
	return []byte(vt.Format(`"` + jsonTime + `"`)), nil
}

//
// projects
//

// ErrCode is a custom type for error codes.
type ErrCode string

// create a list of error codes
const (
	ErrProjectAlreadyExistsCode = "project_already_exists"
	ErrProjectNotFoundCode      = "project_not_found"
)

var mapErrCodeToMessage = map[ErrCode]string{
	ErrProjectAlreadyExistsCode: "project already exists",
	ErrProjectNotFoundCode:      "project not found",
}

// ServiceError is a custom error type.
type ServiceError struct {
	Code ErrCode
	Msg  string
	err  error
}

// Error returns the error message.
func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, mapErrCodeToMessage[e.Code])
}

// Unwrap returns the underlying error.
func (e *ServiceError) Unwrap() error {
	return e.err
}

// NewServiceError creates a new ServiceError.
func NewServiceError(code ErrCode, err error) *ServiceError {
	return &ServiceError{
		Code: code,
		Msg:  mapErrCodeToMessage[code],
		err:  err,
	}
}

// Project represents an individual project.
type Project struct {
	ID          string
	Name        string
	Description string
	CreatedAt   ISOTime
}

//
// SMTP transports
//

// SMTPTransport represents an individual transport based on
type SMTPTransport struct {
	ID            string
	ProjectID     string
	Name          string
	Host          string
	Port          int
	Username      string
	EmailFrom     string
	EmailFromName string
	EmailReplyTo  []string
	CreatedAt     ISOTime
	ModifiedAt    ISOTime
}

// CreateSMTPTransport is the input parameters for the CreateSMTPTransport method.
type CreateSMTPTransport struct {
	ID            string
	ProjectID     string
	Name          string
	Host          string
	Port          int
	Username      string
	Password      string
	EmailFrom     string
	EmailFromName string
	EmailReplyTo  []string
}

//
// groups
//

// Group represents a group of users.
type Group struct {
	ID         string
	ProjectID  string
	Name       string
	CreatedAt  ISOTime
	ModifiedAt ISOTime
}

//
// templates
//

// Template represents a single email template.
type Template struct {
	ID         string
	GroupID    string
	ProjectID  string
	Text       string
	TextDigest string
	HTML       string
	HTMLDigest string
	CreatedAt  ISOTime
	ModifiedAt ISOTime
}

// CreateTemplate is the input parameters for the CreateTemplate method.
type CreateTemplate struct {
	ID         string
	GroupID    string
	ProjectID  string
	Text       string
	TextDigest string
	HTML       string
	HTMLDigest string
}

// CreateTemplateFromFiles is the input parameters for the CreateTemplateFromFiles method.
type CreateTemplateFromFiles struct {
	ID            string
	GroupID       string
	ProjectID     string
	TxtFilenames  []string
	HTMLFilenames []string
}

//
// send email
//

// SendEmailParams is the input parameters for the SendEmail method.
type SendEmailParams struct {
	TemplateID     string
	ProjectID      string
	TransportID    string
	To             []string
	Subject        string
	TemplateParams map[string]string
}
