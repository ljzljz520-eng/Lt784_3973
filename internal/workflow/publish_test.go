package workflow

import (
	"testing"

	"example.com/permission-selector/internal/domain"
)

func TestWorkflowPublish(t *testing.T) {
	f := newFixture(t)
	defer f.database.Close()
	request := f.request(t, "publish-1")
	if _, err := f.selector.Select(domain.SelectionCommand{RequestID: request.ID, Actor: "admin", RefID: "a1", Type: domain.ObjectTypePerson, Name: "Alice"}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := f.selector.Confirm(request.ID, "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.selector.Dispatch(request.ID, "admin", 1, 10); err != nil {
		t.Fatal(err)
	}
	receipt, err := f.service.PublishWorkflow(request.ID, "admin")
	if err != nil || receipt.Request.Status != domain.StatusPublished {
		t.Fatalf("receipt=%+v err=%v", receipt, err)
	}
}
