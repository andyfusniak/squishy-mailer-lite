package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/andyfusniak/squishy-mailer-lite/entity"

	"github.com/andyfusniak/squishy-mailer-lite/service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func run() error {
	const fakeKey string = "a0bf305856098eba7e4bff506021648b"
	svc, err := service.NewEmailService(
		service.WithHexEncodedEncryptionKey(fakeKey),
	)
	if err != nil {
		return err
	}

	// create a new project to test the system
	// if the project already exists then get the existing project
	ctx := context.Background()
	project, err := svc.CreateProject(ctx,
		"the-cloud-project",
		"The Cloud Project",
		"The Cloud Company transactional emails.")
	if err != nil {
		var e *entity.ServiceError
		if errors.As(err, &e) {
			if e.Code == entity.ErrProjectAlreadyExistsCode {
				project, err = svc.GetProject(ctx, "the-cloud-project")
				if err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}

	_, err = svc.CreateSMTPTransport(ctx, entity.CreateSMTPTransport{
		ID:            "the-cloud-transport",
		ProjectID:     project.ID,
		Name:          "Squishy Mailer Lite Transport",
		Host:          "email-smtp.us-east-1.amazonaws.com",
		Port:          587,
		Username:      "<username>",
		Password:      os.Getenv("SQUISHY_MAILER_LITE_SMTP_PASSWORD"),
		EmailFrom:     "support@ravenmailer.com",
		EmailFromName: "Raven Mailer Support",
		EmailReplyTo:  []string{"support@ravenmailer.com"},
	})
	if err != nil {
		return err
	}

	group, err := svc.CreateGroup(ctx, "g1", project.ID, "Group One")
	if err != nil {
		return err
	}

	// _, err = svc.CreateTemplate(ctx, entity.CreateTemplate{
	// 	ID:        "t1",
	// 	ProjectID: project.ID,
	// 	GroupID:   group.ID,
	// 	HTML:      "<h1>Welcome {{.firstname}}, to the Cloud</h1>",
	// 	Text:      "Welcome {{.firstname}}, to the Cloud",
	// })
	// if err != nil {
	// 	return err
	// }

	_, err = svc.SetTemplateFromFiles(ctx, entity.CreateTemplateFromFiles{
		ID:        "t1",
		ProjectID: project.ID,
		GroupID:   group.ID,
		HTMLFilenames: []string{
			"./service/testdata/email/templates/layout.html",
			"./service/testdata/email/templates/welcome.html",
		},
		TxtFilenames: []string{
			"./service/testdata/email/templates/layout.txt",
			"./service/testdata/email/templates/welcome.txt",
		},
	})
	if err != nil {
		return err
	}

	// send a test email
	mq, err := svc.SendEmailAsync(ctx, entity.SendEmailParams{
		TemplateID:  "t1",
		ProjectID:   project.ID,
		TransportID: "the-cloud-transport",
		To:          []string{"andy@andyfusniak.com"},
		Subject:     "My test subject line",
		TemplateParams: map[string]string{
			"firstname": "Andy",
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", mq)

	fmt.Printf("%#v\n", mq.Body)
	fmt.Printf("%#v\n", mq.Metadata)

	return nil
}
