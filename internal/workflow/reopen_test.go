package workflow

import (
	"path/filepath"
	"testing"

	"example.com/permission-selector/internal/audit"
	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/org"
	"example.com/permission-selector/internal/selector"
	"example.com/permission-selector/internal/store"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.db")
	database, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	organization := org.NewService(database)
	auditLog := audit.NewService(database)
	selection := selector.NewService(database, organization, auditLog)
	service := NewService(database, organization, selection, auditLog)
	if _, err := organization.AddDepartment("root", "", "Root"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateRequest(domain.RequestCommand{RequestID: "persist-1", Actor: "admin", Title: "Persisted", Reason: "Reopen should retain data"}); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	request, err := reopened.FindRequest("persist-1")
	if err != nil || request.Title != "Persisted" {
		t.Fatalf("request=%+v err=%v", request, err)
	}
}

func TestWorkflowReopen(t *testing.T) {
	f := newFixture(t)
	defer f.database.Close()
	request := f.request(t, "reopen-1")
	if _, err := f.selector.Select(domain.SelectionCommand{RequestID: request.ID, Actor: "admin", RefID: "root", Type: domain.ObjectTypeDepartment, Name: "Root"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.selector.Confirm(request.ID, "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.selector.Dispatch(request.ID, "admin", 1, 10); err != nil {
		t.Fatal(err)
	}
	snapshot, err := f.service.ReopenAndRead(request.ID)
	if err != nil || len(snapshot.Objects) != 1 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
}
