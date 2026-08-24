package org

import (
	"path/filepath"
	"testing"

	"example.com/permission-selector/internal/store"
)

func TestOrganizationTree(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "org.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := NewService(database)
	if _, err := service.AddDepartment("root", "", "Root"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddDepartment("child", "root", "Child"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddAccount("a1", "child", "alice", "Alice", "alice@example.com"); err != nil {
		t.Fatal(err)
	}
	accounts, err := service.AccountsForNode(structFilter())
	if err != nil || len(accounts) != 1 {
		t.Fatalf("accounts=%+v err=%v", accounts, err)
	}
}
