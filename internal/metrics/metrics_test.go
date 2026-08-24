package metrics

import (
	"path/filepath"
	"testing"

	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/store"
)

func TestBuildReport(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "metrics.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := database.SaveRequest(domain.AuthorizationRequest{ID: "r1", Title: "Report", Applicant: "admin", Status: domain.StatusConfirmed}); err != nil {
		t.Fatal(err)
	}
	report, err := BuildReport(database)
	if err != nil || report.Requests != 1 || report.Confirmed != 1 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}
