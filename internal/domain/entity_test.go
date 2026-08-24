package domain

import "testing"

func TestDomainValidation(t *testing.T) {
	if err := ValidateRequest(AuthorizationRequest{ID: "r1", Title: "Access", Applicant: "admin", Status: StatusDraft}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRequest(AuthorizationRequest{Title: ""}); err == nil {
		t.Fatal("expected invalid request")
	}
	objects := []AuthorizationObject{{Type: ObjectTypePerson}, {Type: ObjectTypeDepartment}, {Type: ObjectTypePerson}}
	if summary := SummarizeObjects(objects); summary.Total != 3 || summary.People != 2 || summary.Departments != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
