package email

import (
	"fmt"
	"net/smtp"

	jemail "github.com/jordan-wright/email"
)

// AWSSMTPTransport sends emails using AWS SES.
type AWSSMTPTransport struct {
	transportName string
	host          string
	port          string
	username      string
	password      string
	name          string
	from          string
}

type AWSConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Name     string
	From     string
}

// NewAWSSMTPTransport creates a new AWS sender.
func NewAWSSMTPTransport(name string, cfg AWSConfig) *AWSSMTPTransport {
	return &AWSSMTPTransport{
		transportName: name,
		host:          cfg.Host,
		port:          cfg.Port,
		username:      cfg.Username,
		password:      cfg.Password,
		name:          cfg.Name,
		from:          cfg.From,
	}
}

// SendEmail sends an email using AWS SES.
func (s *AWSSMTPTransport) SendEmail(params EmailParams) error {
	m := jemail.NewEmail()
	m.From = fmt.Sprintf("%s <%s>", s.name, s.from)
	m.ReplyTo = []string{s.from}
	m.Subject = params.Subject
	m.Text = []byte(params.Text)
	if params.HTML != "" {
		m.HTML = []byte(params.HTML)
	}
	m.To = params.To
	m.Cc = params.Cc
	m.Bcc = params.Bcc
	for _, a := range params.Attachments {
		m.AttachFile(a)
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	return m.Send(addr, auth)
}
