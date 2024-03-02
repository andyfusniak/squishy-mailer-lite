package sqlite3_test

import (
	"context"
	"errors"
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
	assert.WithinDuration(t, time.Now(), obj.CreatedAt.Time, 1*time.Millisecond)
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
		EmailFromName:     "Example Site",
		EmailReplyTo:      store.JSONArray{"reply-to@examplesite.com"},
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
	assert.Equal(t, "Example Site", obj.EmailFromName)
	assert.Equal(t, store.JSONArray{"reply-to@examplesite.com"}, obj.EmailReplyTo)
	assert.WithinDuration(t, time.Now(), obj.CreatedAt.Time, 1*time.Millisecond)
	assert.WithinDuration(t, time.Now(), obj.ModifiedAt.Time, 1*time.Millisecond)
}

func TestInsertGroupIntoNonExistingProject(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	ctx := context.Background()
	group, err := st.InsertGroup(ctx, store.AddGroup{
		GroupID:   "gz",
		ProjectID: "non-existing-project",
		GroupName: "Group Z",
	})
	if err == nil {
		t.Fatalf("expected err to be non-nil")
	}
	if group != nil {
		t.Fatalf("expected group to be nil")
	}

	// assert that error is of type *store.Error and that the code is store.ErrProjectNotFound
	var storeErr *store.Error
	if errors.As(err, &storeErr) {
		if storeErr.Code != store.ErrProjectNotFound {
			t.Fatalf("expected storeErr.Code to be store.ErrProjectNotFound")
		}
	} else {
		t.Fatalf("expected err to be of type *store.Error")
	}
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
	assert.WithinDuration(t, time.Now(), obj.CreatedAt.Time, 1*time.Millisecond)
	assert.WithinDuration(t, time.Now(), obj.ModifiedAt.Time, 1*time.Millisecond)
}

func TestGetGroup(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	ctx := context.Background()
	p1, err := st.InsertProject(ctx, store.AddProject{
		ProjectID:   "p1",
		ProjectName: "Project P One",
		Description: "Project P One Description",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	g1, err := st.InsertGroup(ctx, store.AddGroup{
		GroupID:   "g1",
		ProjectID: p1.ProjectID,
		GroupName: "Group One",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	obj, err := st.GetGroup(ctx, p1.ProjectID, g1.GroupID)
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}
	assert.Equal(t, "g1", obj.GroupID)
	assert.Equal(t, p1.ProjectID, obj.ProjectID)
	assert.Equal(t, "Group One", obj.GroupName)
	assert.Equal(t, g1.CreatedAt, obj.CreatedAt)
	assert.Equal(t, g1.ModifiedAt, obj.ModifiedAt)
}

func TestNonExistentGroupInProject(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	// create project p1
	ctx := context.Background()
	p1, err := st.InsertProject(ctx, store.AddProject{
		ProjectID:   "p1",
		ProjectName: "Project P One",
		Description: "Project P One Description",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	// look for non-existent group in project p1
	g, err := st.GetGroup(ctx, p1.ProjectID, "non-existent-group")
	if err == nil {
		t.Fatalf("expected err to be non-nil")
	}
	assert.Nil(t, g, "expected g to be nil")

	// assert that error is of type *store.Error and that the code is store.ErrGroupNotFound
	var storeErr *store.Error
	if errors.As(err, &storeErr) {
		if storeErr.Code != store.ErrGroupNotFound {
			t.Fatalf("expected storeErr.Code to be store.ErrGroupNotFound")
		}
	} else {
		t.Fatalf("expected err to be of type *store.Error")
	}
}

func TestNonExistentProjectForGroup(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	// create project p1
	ctx := context.Background()
	p1, err := st.InsertProject(ctx, store.AddProject{
		ProjectID:   "p1",
		ProjectName: "Project P One",
		Description: "Project P One Description",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	// create group g1
	g1, err := st.InsertGroup(ctx, store.AddGroup{
		GroupID:   "g1",
		ProjectID: p1.ProjectID,
		GroupName: "Group One",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	// look for group g1 in non-existent project
	_, err = st.GetGroup(ctx, "non-existent-project", g1.GroupID)
	if err == nil {
		t.Fatalf("expected err to be non-nil")
	}

	// assert that the error is of type *store.Error and that the code is store.ErrProjectNotFound
	var storeErr *store.Error
	if errors.As(err, &storeErr) {
		if storeErr.Code != store.ErrProjectNotFound {
			t.Fatalf("expected storeErr.Code to be store.ErrProjectNotFound")
		}
	} else {
		t.Fatalf("expected err to be of type *store.Error")
	}
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
	assert.WithinDuration(t, time.Now(), obj.CreatedAt.Time, 1*time.Millisecond)
	assert.WithinDuration(t, time.Now(), obj.ModifiedAt.Time, 1*time.Millisecond)
}

func TestGetTemplate(t *testing.T) {
	rw, err := setupInMemoryDB()
	if err != nil {
		t.Fatalf("rw, ro, err := openDBs() failed: %v", err)
	}
	defer rw.Close()

	// create a new store
	st := sqlite3.NewStore(rw, rw)

	// create project p1
	ctx := context.Background()
	p1, err := st.InsertProject(ctx, store.AddProject{
		ProjectID:   "p1",
		ProjectName: "P One",
		Description: "P One project description",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	g1, err := st.InsertGroup(ctx, store.AddGroup{
		GroupID:   "g1",
		ProjectID: p1.ProjectID,
		GroupName: "Group One",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	t1, err := st.InsertTemplate(ctx, store.AddTemplate{
		TemplateID: "tmpl1",
		GroupID:    g1.GroupID,
		ProjectID:  p1.ProjectID,
		Txt:        "Test Text",
		HTML:       "<h1>Test HTML</h1>",
	})
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}

	// get template tmpl1 from project p1
	obj, err := st.GetTemplate(ctx, p1.ProjectID, "tmpl1")
	if err != nil {
		t.Fatalf("expected err to be non-nil: %+v", err)
	}
	assert.Equal(t, "tmpl1", obj.TemplateID)
	assert.Equal(t, g1.GroupID, obj.GroupID)
	assert.Equal(t, p1.ProjectID, obj.ProjectID)
	assert.Equal(t, "Test Text", obj.Txt)
	assert.Equal(t, "<h1>Test HTML</h1>", obj.HTML)
	assert.Equal(t, t1.CreatedAt, obj.CreatedAt)
	assert.Equal(t, t1.ModifiedAt, obj.ModifiedAt)

	// get non-existent template from project p1
	obj, err = st.GetTemplate(ctx, p1.ProjectID, "non-existent-template")
	var storeErr *store.Error
	if errors.As(err, &storeErr) {
		if err.(*store.Error).Code != store.ErrTemplateNotFound {
			t.Fatalf("expected err to be store.ErrTemplateNotFound: %+v", err)
		}
	} else {
		t.Fatalf("expected err to be of type *store.Error")
	}

	assert.Nil(t, obj, "expected obj to be nil")

	// get template tmpl1 from non-existent project
	obj, err = st.GetTemplate(ctx, "non-existent-project", "tmpl1")
	if err != nil {
		var storeErr *store.Error
		if errors.As(err, &storeErr) {
			if storeErr.Code != store.ErrProjectNotFound {
				t.Fatalf("expected err to be store.ErrProjectNotFound: %+v", err)
			}
		}
	}
	assert.Nil(t, obj, "expected obj to be nil")
}
