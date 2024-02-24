package service

import (
	"context"
	"time"

	"github.com/andyfusniak/squishy-mailer-lite/entity"
	"github.com/andyfusniak/squishy-mailer-lite/internal/store"
	"github.com/pkg/errors"
)

type Service struct {
	store store.Repository
}

func NewService(store store.Repository) *Service {
	return &Service{
		store: store,
	}
}

//
// projects
//

// CreateProject creates a new project.
func (s *Service) CreateProject(ctx context.Context, id, name, description string) (*entity.Project, error) {
	obj, err := s.store.InsertProject(ctx, store.AddProject{
		ID:          id,
		PName:       name,
		Description: description,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "store.InsertProject failed")
	}
	return projectFromStoreObject(obj), nil
}

func projectFromStoreObject(obj *store.Project) *entity.Project {
	return &entity.Project{
		ID:          obj.ID,
		Name:        obj.PName,
		Description: obj.Description,
		CreatedAt:   entity.ISOTime(obj.CreatedAt),
	}
}

//
// transports
//

// CreateTransport creates a new transport.
func (s *Service) CreateTransport(ctx context.Context, params entity.CreateTransport) (*entity.Transport, error) {
	obj, err := s.store.InsertTransport(ctx, store.AddTransport{
		ID:                params.ID,
		ProjectID:         params.ProjectID,
		TRName:            params.Name,
		Host:              params.Host,
		Port:              params.Port,
		EncryptedPassword: params.Password, // TODO encrypt the password
		Username:          params.Username,
		EmailFrom:         params.EmailFrom,
		EmailReplyTo:      params.EmailReplyTo,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "store.InsertTransport failed")
	}
	return transportFromStoreObject(obj), nil
}

func transportFromStoreObject(obj *store.Transport) *entity.Transport {
	return &entity.Transport{
		ID:           obj.ID,
		ProjectID:    obj.ProjectID,
		Name:         obj.TRName,
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
		ID:         id,
		ProjectID:  projectID,
		GName:      name,
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
		ID:         obj.ID,
		ProjectID:  obj.ProjectID,
		Name:       obj.GName,
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
		ID:         params.ID,
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
		ID:         obj.ID,
		ProjectID:  obj.ProjectID,
		GroupID:    obj.GroupID,
		HTML:       obj.HTML,
		Text:       obj.Txt,
		CreatedAt:  entity.ISOTime(obj.CreatedAt),
		ModifiedAt: entity.ISOTime(obj.ModifiedAt),
	}
}
