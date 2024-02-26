package email

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	gmailSMTPAuthAddr = "smtp.gmail.com"
	gmailSMTPPort     = "587"
)

// GmailSMTPTransport sends emails using Gmail.
type GmailSMTPTransport struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

// NewGmailSMTPTransport creates a new Gmail sender.
func NewGmailSMTPTransport(name, fromEmailAddress, fromEmailPassword string) *GmailSMTPTransport {
	return &GmailSMTPTransport{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

// SendEmail sends an email using Gmail.
func (s *GmailSMTPTransport) SendEmail(params EmailParams) error {
	m := email.NewEmail()
	m.From = fmt.Sprintf("%s <%s>", s.name, s.fromEmailAddress)
	m.ReplyTo = []string{s.fromEmailAddress}
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

	auth := smtp.PlainAuth("", s.fromEmailAddress, s.fromEmailPassword, gmailSMTPAuthAddr)
	addr := fmt.Sprintf("%s:%s", gmailSMTPAuthAddr, gmailSMTPPort)
	return m.Send(addr, auth)
}
