package domain

import "testing"

func TestDomainPage(t *testing.T) {
	accounts := []Account{{ID: "b", Display: "Beta"}, {ID: "a", Display: "Alpha"}, {ID: "c", Display: "Gamma"}}
	page, err := BuildAccountPage(accounts, 1, 2)
	if err != nil || len(page.Items) != 2 || page.Items[0].ID != "a" || !page.HasNext {
		t.Fatalf("unexpected page: %+v err=%v", page, err)
	}
	if _, err := BuildAccountPage(accounts, 3, 2); err == nil {
		t.Fatal("expected invalid page")
	}
}
