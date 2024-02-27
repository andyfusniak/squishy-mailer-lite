package service

import (
	"context"
	"time"

	"github.com/andyfusniak/squishy-mailer-lite/entity"
	"github.com/andyfusniak/squishy-mailer-lite/internal/email"
	"github.com/andyfusniak/squishy-mailer-lite/internal/store"
	"github.com/pkg/errors"
)

type Service struct {
	store     store.Repository
	transport email.Sender
}

// NewEmailService creates a new service with the specified store and sender.
func NewEmailService(store store.Repository, transport email.Sender) *Service {
	return &Service{
		store:     store,
		transport: transport,
	}
}

//
// projects
//

// CreateProject creates a new project.
func (s *Service) CreateProject(ctx context.Context, id, name, description string) (*entity.Project, error) {
	obj, err := s.store.InsertProject(ctx, store.AddProject{
		ProjectID:   id,
		ProjectName: name,
		Description: description,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "store.InsertProject failed")
	}
	return projectFromStoreObject(obj), nil
}

func projectFromStoreObject(obj *store.Project) *entity.Project {
	return &entity.Project{
		ID:          obj.ProjectID,
		Name:        obj.ProjectName,
		Description: obj.Description,
		CreatedAt:   entity.ISOTime(obj.CreatedAt),
	}
}

//
// transports
//

// CreateSMTPTransport creates a new SMTP transport.
func (s *Service) CreateSMTPTransport(ctx context.Context, params entity.CreateSMTPTransport) (*entity.SMTPTransport, error) {
	obj, err := s.store.InsertSMTPTransport(ctx, store.AddSMTPTransport{
		SMTPTransportID:   params.ID,
		ProjectID:         params.ProjectID,
		TransportName:     params.Name,
		Host:              params.Host,
		Port:              params.Port,
		EncryptedPassword: params.Password, // TODO encrypt the password
		Username:          params.Username,
		EmailFrom:         params.EmailFrom,
		EmailReplyTo:      params.EmailReplyTo,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "store.InsertSMTPTransport failed")
	}
	return smtpTransportFromStoreObject(obj), nil
}

func smtpTransportFromStoreObject(obj *store.SMTPTransport) *entity.SMTPTransport {
	return &entity.SMTPTransport{
		ID:           obj.SMTPTransportID,
		ProjectID:    obj.ProjectID,
		Name:         obj.TransportName,
		Host:         obj.Host,
		Port:         obj.Port,
		Username:     obj.Username,
		EmailFrom:    obj.EmailFrom,
		EmailReplyTo: obj.EmailReplyTo,
		CreatedAt:    entity.ISOTime(obj.CreatedAt),
		ModifiedAt:   entity.ISOTime(obj.ModifiedAt),
	}
}

//
// groups
//

// CreateGroup creates a new group. A group is a collection of templates.
// Group id's are unique within a project. A project can have many groups.
func (s *Service) CreateGroup(ctx context.Context, id, projectID, name string) (*entity.Group, error) {
	now := store.Datetime(time.Now().UTC())
	obj, err := s.store.InsertGroup(ctx, store.AddGroup{
		GroupID:    id,
		ProjectID:  projectID,
		GroupName:  name,
		CreatedAt:  now,
		ModifiedAt: now,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "store.InsertGroup failed")
	}
	return groupFromStoreObject(obj), nil
}

func groupFromStoreObject(obj *store.Group) *entity.Group {
	return &entity.Group{
		ID:         obj.GroupID,
		ProjectID:  obj.ProjectID,
		Name:       obj.GroupName,
		CreatedAt:  entity.ISOTime(obj.CreatedAt),
		ModifiedAt: entity.ISOTime(obj.ModifiedAt),
	}
}

//
// templates
//

// CreateTemplate creates a new template. A template is a message template.
// Template id's are unique within a project. A project can have many templates.
// A template belongs to a group. A group can have many templates.
func (s *Service) CreateTemplate(ctx context.Context, params entity.CreateTemplate) (*entity.Template, error) {
	now := store.Datetime(time.Now().UTC())
	obj, err := s.store.InsertTemplate(ctx, store.AddTemplate{
		TemplateID: params.ID,
		ProjectID:  params.ProjectID,
		GroupID:    params.GroupID,
		HTML:       params.HTML,
		Txt:        params.Text,
		CreatedAt:  now,
		ModifiedAt: now,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "store.InsertTemplate failed")
	}
	return templateFromStoreObject(obj), nil
}

func templateFromStoreObject(obj *store.Template) *entity.Template {
	return &entity.Template{
		ID:         obj.TemplateID,
		ProjectID:  obj.ProjectID,
		GroupID:    obj.GroupID,
		HTML:       obj.HTML,
		Text:       obj.Txt,
		CreatedAt:  entity.ISOTime(obj.CreatedAt),
		ModifiedAt: entity.ISOTime(obj.ModifiedAt),
	}
}

// SendEmail sends an email using the specified template.
func (s *Service) SendEmail(ctx context.Context, subject, templateID string) error {
	return s.transport.SendEmail(email.EmailParams{
		Subject: subject,
		Text:    "Hello, World!",
		HTML:    "<h1>Hello, World!</h1>",
		To:      []string{"andy@andyfusniak.com"},
	})
}
