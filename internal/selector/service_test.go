package selector

import (
	"path/filepath"
	"testing"

	"example.com/permission-selector/internal/audit"
	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/org"
	"example.com/permission-selector/internal/store"
)

func TestSelectorAccounts(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "selector.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	organization := org.NewService(database)
	if _, err := organization.AddDepartment("root", "", "Root"); err != nil {
		t.Fatal(err)
	}
	if _, err := organization.AddAccount("a1", "root", "alice", "Alice", "alice@example.com"); err != nil {
		t.Fatal(err)
	}
	request := domain.AuthorizationRequest{ID: "r1", Title: "Access", Applicant: "admin", Status: domain.StatusDraft}
	if err := database.SaveRequest(request); err != nil {
		t.Fatal(err)
	}
	service := NewService(database, organization, audit.NewService(database))
	page, err := service.LoadAccounts(domain.AccountFilter{NodeID: "root"}, 1, 10)
	if err != nil || len(page.Items) != 1 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	if _, err := service.Select(domain.SelectionCommand{RequestID: "r1", Actor: "admin", RefID: "a1", Type: domain.ObjectTypePerson, Name: "Alice"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Select(domain.SelectionCommand{RequestID: "r1", Actor: "admin", RefID: "a1", Type: domain.ObjectTypePerson, Name: "Alice"}); err != domain.ErrDuplicate {
		t.Fatalf("expected duplicate, got %v", err)
	}
}
