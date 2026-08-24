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

type fixture struct {
	database *store.Store
	org      *org.Service
	selector *selector.Service
	service  *Service
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	database, err := store.Open(filepath.Join(t.TempDir(), "workflow.db"))
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
	if _, err := organization.AddDepartment("eng", "root", "Engineering"); err != nil {
		t.Fatal(err)
	}
	if _, err := organization.AddAccount("a1", "eng", "alice", "Alice", "alice@example.com"); err != nil {
		t.Fatal(err)
	}
	return &fixture{database: database, org: organization, selector: selection, service: service}
}

func (f *fixture) request(t *testing.T, id string) domain.AuthorizationRequest {
	t.Helper()
	request, err := f.service.CreateRequest(domain.RequestCommand{RequestID: id, Actor: "admin", Title: "Role access", Reason: "Need access for work"})
	if err != nil {
		t.Fatal(err)
	}
	return request
}
