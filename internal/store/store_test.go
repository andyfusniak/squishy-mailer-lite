package store_test

import (
	"testing"

	"github.com/andyfusniak/squishy-mailer-lite/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestMailQueueSMTPTransportScan(t *testing.T) {
	// Test scan into a value
	var dest store.MailQueueSMTPTransport

	const json string = `
{
  "smtp_transport_id": "tr1",
  "project_id": "p1",
  "transport_name": "Transport One",
  "host": "smtp.example.com",
  "port": 587,
  "username": "user1",
  "encrypted_password": "encrypted_password1",
  "email_from": "support@examplesite.com",
  "email_from_name": "Example Site Support",
  "email_reply_to": ["support@examplesite.com"]
}
`
	// "created_at":  "2024-02-01T22:06:58.678912Z"
	// "modified_at": "2024-03-02T12:30:35.123456Z"
	err := dest.Scan(json)
	if err != nil {
		t.Errorf("dest.Scan() failed: %v", err)
	}

	// make a time.Time from ISO8601 string
	// createdAt, err := time.Parse(time.RFC3339Nano, "2024-02-01T22:06:58.678912Z")
	// if err != nil {
	// 	t.Errorf("time.Parse() failed: %v", err)
	// }
	// modifiedAt, err := time.Parse(time.RFC3339Nano, "2024-03-02T12:30:35.123456Z")
	// if err != nil {
	// 	t.Errorf("time.Parse() failed: %v", err)
	// }

	assert.Equal(t, store.MailQueueSMTPTransport{
		SMTPTransportID:   "tr1",
		ProjectID:         "p1",
		TransportName:     "Transport One",
		Host:              "smtp.example.com",
		Port:              587,
		Username:          "user1",
		EncryptedPassword: "encrypted_password1",
		EmailFrom:         "support@examplesite.com",
		EmailFromName:     "Example Site Support",
		EmailReplyTo:      store.JSONArray{"support@examplesite.com"},
		// CreatedAt:         store.Datetime(createdAt),
		// ModifiedAt:        store.Datetime(modifiedAt),
	}, dest, "scanned value does not match expected value")
}
