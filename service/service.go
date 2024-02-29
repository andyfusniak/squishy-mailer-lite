package service

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"time"

	htmltemplate "html/template"
	txttemplate "text/template"

	"github.com/andyfusniak/squishy-mailer-lite/entity"
	"github.com/andyfusniak/squishy-mailer-lite/internal/email"
	"github.com/andyfusniak/squishy-mailer-lite/internal/secrets"

	"github.com/andyfusniak/squishy-mailer-lite/internal/store"

	"github.com/pkg/errors"
)

type Service struct {
	store         store.Repository
	encryptionKey []byte
}

// NewEmailService creates a new service with the specified store and
// encryption key encKey. The encryption key is used to encrypt and decrypt
// sensitive data such as passwords. It must be 16 bytes in length (128 bits).
// NewEmailService is hard-coded to use AES-GCM with a random nonce.
func NewEmailService(store store.Repository, encKey []byte) *Service {
	return &Service{
		store:         store,
		encryptionKey: encKey,
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
	// encrypt the plaintext password to a hex encoded ciphertext representation.
	// The plaintext password is never stored in the store and the ciphertext
	// is stored in its place.
	mgr, err := secrets.New(secrets.AESGCMWithRandomNonce, s.encryptionKey)
	if err != nil {
		return nil, errors.Wrapf(err, "[service] secrets.New failed")
	}
	nonce, ciphertext, err := mgr.EncryptHexEncode(params.Password)
	if err != nil {
		return nil, errors.Wrapf(err, "[service] mgr.EncryptHexEncode failed")
	}
	encryptedPassword := nonce + ciphertext

	obj, err := s.store.InsertSMTPTransport(ctx, store.AddSMTPTransport{
		SMTPTransportID: params.ID,
		ProjectID:       params.ProjectID,
		TransportName:   params.Name,
		Host:            params.Host,
		Port:            params.Port,
		// hex encoded nonce (12 bytes) + AES GCM encrypted password
		EncryptedPassword: encryptedPassword,
		Username:          params.Username,
		EmailFrom:         params.EmailFrom,
		EmailFromName:     params.EmailFromName,
		EmailReplyTo:      store.JSONArray(params.EmailReplyTo),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "[service] store.InsertSMTPTransport failed")
	}
	return smtpTransportFromStoreObject(obj), nil
}

func (s *Service) GetSMTPTransport(ctx context.Context, transportID, projectID string) (*entity.SMTPTransport, error) {
	obj, err := s.store.GetSMTPTransport(ctx, transportID, projectID)
	if err != nil {
		return nil, errors.Wrapf(err, "[service] store.GetSMTPTransport failed")
	}
	return smtpTransportFromStoreObject(obj), nil
}

func smtpTransportFromStoreObject(obj *store.SMTPTransport) *entity.SMTPTransport {
	return &entity.SMTPTransport{
		ID:            obj.SMTPTransportID,
		ProjectID:     obj.ProjectID,
		Name:          obj.TransportName,
		Host:          obj.Host,
		Port:          obj.Port,
		Username:      obj.Username,
		EmailFrom:     obj.EmailFrom,
		EmailFromName: obj.EmailFromName,
		EmailReplyTo:  obj.EmailReplyTo,
		CreatedAt:     entity.ISOTime(obj.CreatedAt),
		ModifiedAt:    entity.ISOTime(obj.ModifiedAt),
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
		Txt:        params.Text,
		TxtDigest:  params.TextDigest,
		HTML:       params.HTML,
		HTMLDigest: params.HTMLDigest,
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
		Text:       obj.Txt,
		TextDigest: obj.TxtDigest,
		HTML:       obj.HTML,
		HTMLDigest: obj.HTMLDigest,
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

func amalgalateTemplates(filenames []string) ([]byte, error) {
	// concat the filenames into a buffer
	var buf bytes.Buffer

	for _, f := range filenames {
		// read the file into a string
		// and append it to the txt and html strings
		content, err := os.ReadFile(f)
		if err != nil {
			return nil, errors.Wrapf(err, "[service] os.ReadFile failed")
		}
		_, err = buf.Write(content)
		if err != nil {
			return nil, errors.Wrapf(err, "[service] sbTxt.Write failed")
		}
	}

	return buf.Bytes(), nil
}

// CreateTemplateFromFiles creates a new template from the specified files.
func (s *Service) CreateTemplateFromFiles(ctx context.Context, params entity.CreateTemplateFromFiles) (*entity.Template, error) {
	// txt templates
	if err := checkTemplates(txtTemplate, params.TxtFilenames...); err != nil {
		return nil, errors.Wrapf(err, "[service] checkTemplates txt failed")
	}
	// amalgalate the txt templates into a single string
	txt, err := amalgalateTemplates(params.TxtFilenames)
	if err != nil {
		return nil, errors.Wrapf(err, "[service] amalgalateTemplates txt failed")
	}
	// create a SHA512 (224 bit) hash of the text template amalgalated string
	hash := sha512.New512_224()
	hash.Write(txt)
	sum := hash.Sum(nil)
	txtCS := hex.EncodeToString(sum[0:16])

	// html templates
	if err := checkTemplates(htmlTemplate, params.HTMLFilenames...); err != nil {
		return nil, errors.Wrapf(err, "[service] checkTemplates html failed")
	}
	// amalgalate the html templates into a single string
	html, err := amalgalateTemplates(params.HTMLFilenames)
	if err != nil {
		return nil, errors.Wrapf(err, "[service] amalgalateTemplates html failed")
	}
	// create a SHA512 (224 bit) hash of the html template amalgalated string
	hash = sha512.New512_224()
	hash.Write(html)
	sum = hash.Sum(nil)
	htmlCS := hex.EncodeToString(sum[0:16])

	return s.CreateTemplate(ctx, entity.CreateTemplate{
		ID:         params.ID,
		ProjectID:  params.ProjectID,
		GroupID:    params.GroupID,
		Text:       string(txt),
		TextDigest: txtCS,
		HTML:       string(html),
		HTMLDigest: htmlCS,
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

	trObj, err := s.store.GetSMTPTransport(ctx, params.TransportID, params.ProjectID)
	if err != nil {
		return errors.Wrapf(err, "[service] store.GetSMTPTransport failed")
	}

	// decrypt the password
	mgr, err := secrets.New(secrets.AESGCMWithRandomNonce, s.encryptionKey)
	if err != nil {
		return err
	}
	pwPlaintext, err := mgr.HexDecodeDecrypt(trObj.EncryptedPassword[:24], trObj.EncryptedPassword[24:])
	if err != nil {
		return err
	}

	awsTransport := email.NewAWSSMTPTransport(email.AWSConfig{
		Host:     trObj.Host,
		Port:     trObj.Port,
		Username: trObj.Username,
		Password: pwPlaintext,
		From:     trObj.EmailFrom,
		FromName: trObj.EmailFromName,
		ReplyTo:  trObj.EmailReplyTo,
	})

	return awsTransport.SendEmail(email.EmailParams{
		Subject: params.Subject,
		Text:    txt.String(),
		HTML:    html.String(),
		To:      params.To,
	})
}
