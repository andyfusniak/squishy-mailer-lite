package email

type Sender interface {
	SendEmail(params EmailParams) error
}

// EmailParams are the parameters for sending an email.
type EmailParams struct {
	// Subject is the subject of the email
	Subject string

	// Text and HTML are the body of the email
	Text string
	HTML string

	// From optional override for default sender
	From string

	// To, Cc, Bcc are the recipients of the email
	To  []string
	Cc  []string
	Bcc []string

	// Attachments are the files to attach to the email
	Attachments []string
}
