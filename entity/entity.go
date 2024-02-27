package entity

import "time"

const jsonTime = "2006-01-02T15:04:05.000Z07:00" // .000Z = keep trailing zeros

// ISOTime custom type to allow for JSON microsecond formating.
type ISOTime time.Time

// MarshalJSON provides microsecond formating
func (t ISOTime) MarshalJSON() ([]byte, error) {
	vt := time.Time(t)
	vt = vt.UTC().Round(time.Millisecond)
	return []byte(vt.Format(`"` + jsonTime + `"`)), nil
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
	ID           string
	ProjectID    string
	Name         string
	Host         string
	Port         int
	Username     string
	EmailFrom    string
	EmailReplyTo string
	CreatedAt    ISOTime
	ModifiedAt   ISOTime
}

// CreateSMTPTransport is the input parameters for the CreateSMTPTransport method.
type CreateSMTPTransport struct {
	ID           string
	ProjectID    string
	Name         string
	Host         string
	Port         int
	Username     string
	Password     string
	EmailFrom    string
	EmailReplyTo string
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
	HTML       string
	Text       string
	CreatedAt  ISOTime
	ModifiedAt ISOTime
}

// CreateTemplate is the input parameters for the CreateTemplate method.
type CreateTemplate struct {
	ID        string
	GroupID   string
	ProjectID string
	HTML      string
	Text      string
}

//
// send email
//

// SendEmailParams is the input parameters for the SendEmail method.
type SendEmailParams struct {
	ProjectID      string
	TemplateID     string
	To             []string
	Subject        string
	TemplateParams map[string]string
}
