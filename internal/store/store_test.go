package store

import (
	"path/filepath"
	"testing"
	"time"

	"example.com/permission-selector/internal/domain"
)

func TestStoreRoundTrip(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "selector.db"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	node := domain.OrganizationNode{ID: "root", Name: "Root", Path: "root", Active: true, CreatedAt: now, UpdatedAt: now}
	if err := database.SaveNode(node); err != nil {
		t.Fatal(err)
	}
	got, err := database.FindNode("root")
	if err != nil || got.Name != "Root" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
}
