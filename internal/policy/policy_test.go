package policy

import (
	"testing"

	"example.com/permission-selector/internal/domain"
)

func TestPolicyDecisions(t *testing.T) {
	engine := NewEngine()
	if !engine.CanCreate("admin").Allowed {
		t.Fatal("admin should create")
	}
	if engine.CanPublish("guest").Allowed {
		t.Fatal("guest should not publish")
	}
	if err := engine.ValidateTransition("admin", domain.StatusDraft, domain.StatusConfirmed, 1); err != nil {
		t.Fatal(err)
	}
}
