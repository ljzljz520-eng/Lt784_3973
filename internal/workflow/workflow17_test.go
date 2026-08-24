package workflow

import (
	"testing"

	"example.com/permission-selector/internal/domain"
)

func TestWorkflow17(t *testing.T) {
	f := newFixture(t)
	defer f.database.Close()
	request := f.request(t, "regression-17")
	if _, err := f.selector.Select(domain.SelectionCommand{RequestID: request.ID, Actor: "admin", RefID: "a1", Type: domain.ObjectTypePerson, Name: "Alice"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.selector.Confirm(request.ID, "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.selector.Dispatch(request.ID, "admin", 1, 1); err != nil {
		t.Fatal(err)
	}
	page, err := f.selector.ResultPage(request.ID, 1, 1)
	if err != nil {
		t.Fatalf("result page should be readable after first dispatch: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].RefID != "a1" {
		t.Fatalf("unexpected result page: %+v", page)
	}
}
