package api

import (
	"net/http/httptest"
	"path/filepath"
	"testing"

	"example.com/permission-selector/internal/audit"
	"example.com/permission-selector/internal/config"
	"example.com/permission-selector/internal/org"
	"example.com/permission-selector/internal/selector"
	"example.com/permission-selector/internal/store"
	"example.com/permission-selector/internal/workflow"
)

func TestHealthEndpoint(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	organization := org.NewService(database)
	auditLog := audit.NewService(database)
	selection := selector.NewService(database, organization, auditLog)
	server := NewServer(config.Config{HTTPAddress: ":0", PageSize: 10}, database, organization, selection, workflow.NewService(database, organization, selection, auditLog))
	recorder := httptest.NewRecorder()
	server.health(recorder, httptest.NewRequest("GET", "/health", nil))
	if recorder.Code != 200 {
		t.Fatalf("status=%d", recorder.Code)
	}
}
