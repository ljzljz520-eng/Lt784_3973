package report

import (
	"path/filepath"
	"testing"

	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/store"
)

func TestReportExport(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "report.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.SaveRequest(domain.AuthorizationRequest{ID: "r1", Title: "Export", Applicant: "admin", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	data, err := CSV(database, "r1")
	if err != nil || len(data) < 20 {
		t.Fatalf("len=%d err=%v", len(data), err)
	}
}
