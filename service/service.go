package service

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	htmltemplate "html/template"
	txttemplate "text/template"

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
		return nil, errors.Wrapf(err, "[service] store.InsertProject failed")
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
		return nil, errors.Wrapf(err, "[service] store.InsertSMTPTransport failed")
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
		return nil, errors.Wrapf(err, "[service] store.InsertGroup failed")
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

// CreateTemplate creates a new template using text and HTML strings.
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
		return nil, errors.Wrapf(err, "[service] store.InsertTemplate failed")
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

type templateType int

const (
	txtTemplate templateType = iota
	htmlTemplate
)

func checkTemplates(mode templateType, filenames ...string) error {
	if mode == txtTemplate {
		tmpl, err := txttemplate.ParseFiles(filenames...)
		if err != nil {
			return errors.Wrapf(err, "[service] txt template.ParseFiles failed")
		}

		// write the template to /dev/null to check for errors
		if err := tmpl.ExecuteTemplate(io.Discard, "layout", nil); err != nil {
			return errors.Wrapf(err, "[service] txt tmpl.ExecuteTemplate failed")
		}
	} else {
		tmpl, err := htmltemplate.ParseFiles(filenames...)
		if err != nil {
			return errors.Wrapf(err, "[service] html template.ParseFiles failed")
		}

		// write the template to /dev/null to check for errors
		if err := tmpl.ExecuteTemplate(io.Discard, "layout", nil); err != nil {
			return errors.Wrapf(err, "[service] html tmpl.ExecuteTemplate failed")
		}
	}

	return nil
}

func amalgalateTemplates(filenames []string) (string, error) {
	// concat the filenames into a single string
	var sb strings.Builder
	for _, f := range filenames {
		// read the file into a string
		// and append it to the txt and html strings
		content, err := os.ReadFile(f)
		if err != nil {
			return "", errors.Wrapf(err, "[service] os.ReadFile failed")
		}
		_, err = sb.Write(content)
		if err != nil {
			return "", errors.Wrapf(err, "[service] sbTxt.Write failed")
		}
	}

	return sb.String(), nil
}

// CreateTemplateFromFiles creates a new template from the specified files.
func (s *Service) CreateTemplateFromFiles(ctx context.Context, params entity.CreateTemplateFromFiles) (*entity.Template, error) {
	// check the txt and html templates for errors
	if err := checkTemplates(txtTemplate, params.TxtFilenames...); err != nil {
		return nil, errors.Wrapf(err, "[service] checkTemplates txt failed")
	}
	if err := checkTemplates(htmlTemplate, params.HTMLFilenames...); err != nil {
		return nil, errors.Wrapf(err, "[service] checkTemplates html failed")
	}

	// amalgalate the txt and html templates into a single string
	txt, err := amalgalateTemplates(params.TxtFilenames)
	if err != nil {
		return nil, errors.Wrapf(err, "[service] amalgalateTemplates txt failed")
	}
	html, err := amalgalateTemplates(params.HTMLFilenames)
	if err != nil {
		return nil, errors.Wrapf(err, "[service] amalgalateTemplates html failed")
	}

	// create the template
	return s.CreateTemplate(ctx, entity.CreateTemplate{
		ID:        params.ID,
		ProjectID: params.ProjectID,
		GroupID:   params.GroupID,
		HTML:      html,
		Text:      txt,
	})
}

// SendEmail sends an email using the specified template.
func (s *Service) SendEmail(ctx context.Context, params entity.SendEmailParams) error {
	// retrieve the template from the store
	t, err := s.store.GetTemplate(ctx, params.ProjectID, params.TemplateID)
	if err != nil {
		return errors.Wrapf(err, "[service] store.GetTemplate failed")
	}

	// parse the template string using go text/template
	// and execute the template to produce the final email body
	// and subject
	textTmpl, err := txttemplate.New("layout").Parse(t.Txt)
	if err != nil {
		return errors.Wrapf(err, "[service] txt template.New.Parse failed")
	}
	var txt strings.Builder
	if err := textTmpl.ExecuteTemplate(&txt, "layout", params.TemplateParams); err != nil {
		return errors.Wrapf(err, "[service] txt tmpl.ExecuteTemplate failed")
	}

	htmlTmpl, err := htmltemplate.New("layout").Parse(t.HTML)
	if err != nil {
		return errors.Wrapf(err, "[service] html template.New.Parse failed")
	}
	var html strings.Builder
	if err := htmlTmpl.ExecuteTemplate(&html, "layout", params.TemplateParams); err != nil {
		return errors.Wrapf(err, "[service] html tmpl.ExecuteTemplate failed")
	}

	return s.transport.SendEmail(email.EmailParams{
		Subject: params.Subject,
		Text:    txt.String(),
		HTML:    html.String(),
		To:      params.To,
	})
}
