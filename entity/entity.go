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
// transports
//

// Transport represents an individual transport based on
type Transport struct {
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

// CreateTransport is the input parameters for the CreateTransport method.
type CreateTransport struct {
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
