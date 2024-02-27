package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/andyfusniak/squishy-mailer-lite/entity"

	"github.com/andyfusniak/squishy-mailer-lite/internal/email"
	"github.com/andyfusniak/squishy-mailer-lite/internal/store/sqlite3"
	"github.com/andyfusniak/squishy-mailer-lite/service"
)

const (
	defaultMaxOpenConns int = 120
	defaultMaxIdleConns int = 20
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	// get the database file path from the environment
	dbfile := os.Getenv("DB_FILEPATH")
	if dbfile == "" {
		fmt.Fprint(os.Stderr, "DB_FILEPATH not set\n")
		os.Exit(1)
	}

	// if the dbfile does not exist, create it
	var createDB bool
	if _, err := os.Stat(dbfile); os.IsNotExist(err) {
		createDB = true
	}

	// database connection
	// one read-only with high concurrency
	// one read-write for non-concurrent queries
	rw, err := sqlite3.OpenDB(dbfile)
	if err != nil {
		return err
	}
	defer rw.Close()
	rw.SetMaxOpenConns(1)
	rw.SetMaxIdleConns(1)
	rw.SetConnMaxIdleTime(5 * time.Minute)

	ro, err := sqlite3.OpenDB(dbfile)
	if err != nil {
		return err
	}
	defer ro.Close()
	ro.SetMaxOpenConns(defaultMaxOpenConns)
	ro.SetMaxIdleConns(defaultMaxIdleConns)
	ro.SetConnMaxIdleTime(5 * time.Minute)

	// if the database file did not exist, create the schema
	if createDB {
		if err := sqlite3.CreateSqliteDBSchema(rw); err != nil {
			return fmt.Errorf("failed to create database schema: %w", err)
		}
	}

	// create the store and service
	st := sqlite3.NewStore(rw, rw)

	awsTransport := email.NewAWSSMTPTransport("transport1", email.AWSConfig{
		Host:     "email-smtp.us-east-1.amazonaws.com",
		Port:     "587",
		Username: "<username>",
		Password: "<password>",
		Name:     "Squishy Mailer Lite",
		From:     "support@ravenmailer.com",
	})

	svc := service.NewEmailService(st, awsTransport)

	// create a new project to test the system
	ctx := context.Background()
	project, err := svc.CreateProject(ctx,
		"the-cloud-project",
		"The Cloud Project",
		"The Cloud Company transactional emails.")
	if err != nil {
		return err
	}
	fmt.Printf("project: %+v\n", project)

	smtp, err := svc.CreateSMTPTransport(ctx, entity.CreateSMTPTransport{
		ID:           "the-cloud-transport",
		ProjectID:    project.ID,
		Name:         "The Cloud SMTP",
		Host:         "smtp.sendgrid.net",
		Port:         587,
		Username:     "example",
		Password:     "secret",
		EmailFrom:    "info@example.com",
		EmailReplyTo: "info@example.com",
	})
	if err != nil {
		return err
	}
	fmt.Printf("transport: %#v\n", smtp)

	group, err := svc.CreateGroup(ctx, "g1", project.ID, "Group One")
	if err != nil {
		return err
	}
	fmt.Printf("group: %#v\n", group)

	// template, err := svc.CreateTemplate(ctx, entity.CreateTemplate{
	// 	ID:        "t1",
	// 	ProjectID: project.ID,
	// 	GroupID:   group.ID,
	// 	HTML:      "<h1>Welcome {{.firstname}}, to the Cloud</h1>",
	// 	Text:      "Welcome {{.firstname}}, to the Cloud",
	// })
	// if err != nil {
	// 	return err
	// }

	template, err := svc.CreateTemplateFromFiles(ctx, entity.CreateTemplateFromFiles{
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
	fmt.Printf("template: %#v\n", template)

	// send a test email
	if err := svc.SendEmail(ctx, entity.SendEmailParams{
		TemplateID: template.ID,
		ProjectID:  project.ID,
		To:         []string{"andy@andyfusniak.com"},
		Subject:    "My test subject line",
		TemplateParams: map[string]string{
			"firstname": "Andy",
		},
	}); err != nil {
		return err
	}

	return nil
}
