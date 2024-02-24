package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"

	driversqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"

	"github.com/golang-migrate/migrate/v4/source/httpfs"

	"github.com/andyfusniak/squishy-mailer-lite/entity"
	"github.com/andyfusniak/squishy-mailer-lite/internal/store/sqlite3"
	"github.com/andyfusniak/squishy-mailer-lite/internal/store/sqlite3/schema"
	"github.com/andyfusniak/squishy-mailer-lite/service"
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

	// setup the database connection
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

	// if the database file did not exist, create the schema
	if createDB {
		if err := createSqliteDBSchema(rw); err != nil {
			return fmt.Errorf("failed to create database schema: %w", err)
		}
	}

	// create the store and service
	st := sqlite3.NewStore(rw, rw)
	svc := service.NewService(st)

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

	transport, err := svc.CreateTransport(ctx, entity.CreateTransport{
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
	fmt.Printf("transport: %#v\n", transport)

	group, err := svc.CreateGroup(ctx, "g1", project.ID, "Group One")
	if err != nil {
		return err
	}
	fmt.Printf("group: %#v\n", group)

	template, err := svc.CreateTemplate(ctx, entity.CreateTemplate{
		ID:        "t1",
		ProjectID: project.ID,
		GroupID:   group.ID,
		HTML:      "<h1>Welcome to the Cloud</h1>",
		Text:      "Welcome to the Cloud",
	})
	if err != nil {
		return err
	}
	fmt.Printf("template: %#v\n", template)

	return nil
}

func createSqliteDBSchema(db *sql.DB) error {
	driver, err := driversqlite3.WithInstance(db, &driversqlite3.Config{NoTxWrap: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed with instance %+v\n", err)
		os.Exit(1)
	}

	source, err := httpfs.New(http.FS(schema.Migrations), "migrations")
	if err != nil {
		return err
	}

	mg, err := migrate.NewWithInstance("https", source, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to get new migrate instance: %w", err)
	}

	if err := mg.Up(); err != nil {
		return fmt.Errorf("migrate up failed: %w", err)
	}

	return nil
}
