package workflow

import (
	"testing"

	"example.com/permission-selector/internal/domain"
)

func TestWorkflowAccept(t *testing.T) {
	f := newFixture(t)
	defer f.database.Close()
	receipt, err := f.service.Accept(domain.RequestCommand{RequestID: "accept-1", Actor: "admin", Title: "Engineering access", Reason: "Project team needs access"})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Request.Status != domain.StatusDispatched || receipt.Summary.Total != 1 || receipt.Dispatch.ObjectCount != 1 {
		t.Fatalf("unexpected receipt: %+v", receipt)
	}
}
