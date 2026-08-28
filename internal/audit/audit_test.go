package audit

import (
	"path/filepath"
	"testing"

	"example.com/permission-selector/internal/store"
)

func TestAuditHistory(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := NewService(database)
	if err := service.Write("r1", "admin", "create", "started"); err != nil {
		t.Fatal(err)
	}
	if count, err := service.Count("r1"); err != nil || count != 1 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}
