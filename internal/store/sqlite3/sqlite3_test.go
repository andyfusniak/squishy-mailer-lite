package sqlite3_test

import (
	"context"
	"log"
	"time"

	"github.com/andyfusniak/squishy-mailer-lite/internal/store"
	"github.com/andyfusniak/squishy-mailer-lite/internal/store/sqlite3"
	"github.com/stretchr/testify/assert"

	"database/sql"
	"testing"
)

func setupInMemoryDB() (rw *sql.DB, err error) {
	// database connections
	// one read-write for non-concurrent queries
	rw, err = sqlite3.OpenDB(":memory:")
	if err != nil {
		return nil, err
	}
	rw.SetMaxOpenConns(1)
	rw.SetMaxIdleConns(1)
	rw.SetConnMaxIdleTime(5 * time.Minute)

	if err := sqlite3.CreateSqliteDBSchema(rw); err != nil {
		log.Fatal("failed to create database schema")
	}

	return rw, nil
}

// TestInsertProject tests the InsertProject method with an in-memory
// sqlite3 database. The test creates a new project and checks that
// the returned object is non-nil and that all the fields match the
// inserted input. It also checks that the created and modified
// timestamps are very close to now.
func TestInsertProject(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	ctx := context.Background()
	obj, err := st.InsertProject(ctx, store.AddProject{
		ProjectID:   "test-project",
		ProjectName: "Test Project",
		Description: "A test project",
	})
	if err != nil {
		t.Errorf("expected err to be non-nil: %+v", err)
	}

	// check that the returned object is nil and
	// that all the fields match the inserted input.
	if obj == nil {
		t.Fatalf("expected obj to be non-nil")
	}
	assert.Equal(t, "test-project", obj.ProjectID)
	assert.Equal(t, "Test Project", obj.ProjectName)
	assert.Equal(t, "A test project", obj.Description)

	// check created and modified timestamps are very close to now
	// as we can't know the exact time it was created.
	assert.WithinDuration(t, time.Now(), time.Time(obj.CreatedAt), 1*time.Millisecond)
}

func TestInsertSMTPTransport(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	ctx := context.Background()
	projectObj, err := st.InsertProject(ctx, store.AddProject{
		ProjectID:   "test-project",
		ProjectName: "Test Project",
		Description: "A test project",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	obj, err := st.InsertSMTPTransport(ctx, store.AddSMTPTransport{
		SMTPTransportID:   "test-transport-1",
		ProjectID:         projectObj.ProjectID,
		TransportName:     "Test Transport One",
		Host:              "email-smtp.us-east-1.amazonaws.com",
		Port:              587,
		Username:          "someuser",
		EncryptedPassword: "encryptedpassword",
		EmailFrom:         "from@examplesite.com",
		EmailReplyTo:      "reply-to@examplesite.com",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}
	assert.Equal(t, "test-transport-1", obj.SMTPTransportID)
	assert.Equal(t, projectObj.ProjectID, obj.ProjectID)
	assert.Equal(t, "Test Transport One", obj.TransportName)
	assert.Equal(t, "email-smtp.us-east-1.amazonaws.com", obj.Host)
	assert.Equal(t, 587, obj.Port)
	assert.Equal(t, "someuser", obj.Username)
	assert.Equal(t, "encryptedpassword", obj.EncryptedPassword)
	assert.Equal(t, "from@examplesite.com", obj.EmailFrom)
	assert.Equal(t, "reply-to@examplesite.com", obj.EmailReplyTo)
	assert.WithinDuration(t, time.Now(), time.Time(obj.CreatedAt), 1*time.Millisecond)
	assert.WithinDuration(t, time.Now(), time.Time(obj.ModifiedAt), 1*time.Millisecond)
}

func TestInsertGroup(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	ctx := context.Background()
	projectObj, err := st.InsertProject(ctx, store.AddProject{
		ProjectID:   "test-project",
		ProjectName: "Test Project",
		Description: "A test project",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	obj, err := st.InsertGroup(ctx, store.AddGroup{
		GroupID:   "test-group-1",
		ProjectID: projectObj.ProjectID,
		GroupName: "Test Group One",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}
	assert.Equal(t, "test-group-1", obj.GroupID)
	assert.Equal(t, projectObj.ProjectID, obj.ProjectID)
	assert.Equal(t, "Test Group One", obj.GroupName)
	assert.WithinDuration(t, time.Now(), time.Time(obj.CreatedAt), 1*time.Millisecond)
	assert.WithinDuration(t, time.Now(), time.Time(obj.ModifiedAt), 1*time.Millisecond)
}

func TestInsertTemplate(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	ctx := context.Background()
	projectObj, err := st.InsertProject(ctx, store.AddProject{
		ProjectID:   "p1",
		ProjectName: "P One",
		Description: "P One project description",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	groupObj, err := st.InsertGroup(ctx, store.AddGroup{
		GroupID:   "g1",
		ProjectID: projectObj.ProjectID,
		GroupName: "Group One",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	obj, err := st.InsertTemplate(ctx, store.AddTemplate{
		TemplateID: "tmpl1",
		GroupID:    groupObj.GroupID,
		ProjectID:  groupObj.ProjectID,
		Txt:        "Test Text",
		HTML:       "<h1>Test HTML</h1>",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}
	assert.Equal(t, "tmpl1", obj.TemplateID)
	assert.Equal(t, obj.GroupID, groupObj.GroupID)
	assert.Equal(t, obj.ProjectID, projectObj.ProjectID)
	assert.Equal(t, "Test Text", obj.Txt)
	assert.Equal(t, "<h1>Test HTML</h1>", obj.HTML)
	assert.WithinDuration(t, time.Now(), time.Time(obj.CreatedAt), 1*time.Millisecond)
	assert.WithinDuration(t, time.Now(), time.Time(obj.ModifiedAt), 1*time.Millisecond)
}
