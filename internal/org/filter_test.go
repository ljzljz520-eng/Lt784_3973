package org

import (
	"testing"

	"example.com/permission-selector/internal/domain"
)

func structFilter() domain.AccountFilter {
	return domain.AccountFilter{NodeID: "root", OnlyActive: true}
}

func TestOrganizationFilter(t *testing.T) {
	filter := domain.NormalizeFilter(domain.AccountFilter{})
	if filter.NodeID != "root" || !filter.OnlyActive {
		t.Fatalf("unexpected filter: %+v", filter)
	}
}
