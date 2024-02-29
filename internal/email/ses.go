package email

import (
	"fmt"
	"net/smtp"

	jemail "github.com/jordan-wright/email"
)

// AWSSMTPTransport sends emails using AWS SES.
type AWSSMTPTransport struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
	replyTo  []string
}

type AWSConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	ReplyTo  []string
}

// NewAWSSMTPTransport creates a new AWS sender.
func NewAWSSMTPTransport(cfg AWSConfig) *AWSSMTPTransport {
	return &AWSSMTPTransport{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		fromName: cfg.FromName,
	}
}

// SendEmail sends an email using AWS SES.
func (s *AWSSMTPTransport) SendEmail(params EmailParams) error {
	m := jemail.NewEmail()
	m.From = fmt.Sprintf("%s <%s>", s.fromName, s.from)
	m.ReplyTo = s.replyTo
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
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	return m.Send(addr, auth)
}
