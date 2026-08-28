package policy

import (
	"sort"
	"strings"

	"example.com/permission-selector/internal/domain"
)

type Review struct {
	Actor         string   `json:"actor"`
	Role          Role     `json:"role"`
	ObjectCount   int      `json:"object_count"`
	HasDepartment bool     `json:"has_department"`
	HasPerson     bool     `json:"has_person"`
	Ready         bool     `json:"ready"`
	Reasons       []string `json:"reasons"`
}

func (e *Engine) Review(actor string, objects []domain.AuthorizationObject) Review {
	review := Review{Actor: strings.TrimSpace(actor), Role: e.ResolveRole(actor), ObjectCount: len(objects), Reasons: make([]string, 0)}
	for _, object := range objects {
		switch object.Type {
		case domain.ObjectTypeDepartment:
			review.HasDepartment = true
		case domain.ObjectTypePerson:
			review.HasPerson = true
		default:
			review.Reasons = append(review.Reasons, "unsupported object type")
		}
		if strings.TrimSpace(object.Name) == "" {
			review.Reasons = append(review.Reasons, "object name is blank")
		}
	}
	if len(objects) == 0 {
		review.Reasons = append(review.Reasons, "no objects selected")
	}
	if e.ruleFor(actor).MaximumSelection < len(objects) {
		review.Reasons = append(review.Reasons, "selection exceeds maximum")
	}
	if len(review.Reasons) == 0 {
		review.Ready = true
	}
	return review
}

func (e *Engine) AllowedTypes(actor string) []domain.ObjectType {
	rule := e.ruleFor(actor)
	result := make([]domain.ObjectType, 0, len(rule.AllowedTypes))
	for objectType, allowed := range rule.AllowedTypes {
		if allowed {
			result = append(result, objectType)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (e *Engine) ValidateTransition(actor string, from, to domain.Status, count int) error {
	switch to {
	case domain.StatusConfirmed:
		if err := Require(e.CanConfirm(actor, count)); err != nil {
			return err
		}
		if from != domain.StatusDraft {
			return domain.ErrConflict
		}
	case domain.StatusDispatched:
		if err := Require(e.CanDispatch(actor)); err != nil {
			return err
		}
		if from != domain.StatusConfirmed && from != domain.StatusDispatched {
			return domain.ErrConflict
		}
	case domain.StatusPublished:
		if err := Require(e.CanPublish(actor)); err != nil {
			return err
		}
		if from != domain.StatusDispatched {
			return domain.ErrConflict
		}
	default:
		return domain.ErrInvalidInput
	}
	return nil
}
