package policy

import (
	"fmt"
	"strings"

	"example.com/permission-selector/internal/domain"
)

type Role string

const (
	RoleAdministrator Role = "administrator"
	RoleReviewer      Role = "reviewer"
	RoleRequester     Role = "requester"
)

type Rule struct {
	Role             Role
	CanCreate        bool
	CanSelect        bool
	CanConfirm       bool
	CanDispatch      bool
	CanPublish       bool
	AllowedTypes     map[domain.ObjectType]bool
	MaximumSelection int
}

type Decision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
	Role    Role   `json:"role"`
}

type Engine struct {
	rules map[Role]Rule
}

func NewEngine() *Engine {
	return &Engine{rules: map[Role]Rule{
		RoleAdministrator: {Role: RoleAdministrator, CanCreate: true, CanSelect: true, CanConfirm: true, CanDispatch: true, CanPublish: true, AllowedTypes: map[domain.ObjectType]bool{domain.ObjectTypeDepartment: true, domain.ObjectTypePerson: true}, MaximumSelection: 500},
		RoleReviewer:      {Role: RoleReviewer, CanCreate: false, CanSelect: false, CanConfirm: true, CanDispatch: true, CanPublish: true, AllowedTypes: map[domain.ObjectType]bool{domain.ObjectTypeDepartment: true, domain.ObjectTypePerson: true}, MaximumSelection: 500},
		RoleRequester:     {Role: RoleRequester, CanCreate: true, CanSelect: true, CanConfirm: false, CanDispatch: false, CanPublish: false, AllowedTypes: map[domain.ObjectType]bool{domain.ObjectTypePerson: true}, MaximumSelection: 50},
	}}
}

func (e *Engine) ResolveRole(actor string) Role {
	switch strings.ToLower(strings.TrimSpace(actor)) {
	case "admin", "administrator", "root":
		return RoleAdministrator
	case "reviewer", "auditor":
		return RoleReviewer
	default:
		return RoleRequester
	}
}

func (e *Engine) ruleFor(actor string) Rule {
	role := e.ResolveRole(actor)
	return e.rules[role]
}

func (e *Engine) CanCreate(actor string) Decision {
	rule := e.ruleFor(actor)
	return decision(rule.Role, rule.CanCreate, "actor cannot create authorization requests")
}

func (e *Engine) CanSelect(actor string, objectType domain.ObjectType, current int) Decision {
	rule := e.ruleFor(actor)
	if !rule.CanSelect {
		return decision(rule.Role, false, "actor cannot select authorization objects")
	}
	if !rule.AllowedTypes[objectType] {
		return decision(rule.Role, false, "object type is not allowed for actor")
	}
	if current >= rule.MaximumSelection {
		return decision(rule.Role, false, "selection limit has been reached")
	}
	return decision(rule.Role, true, "selection allowed")
}

func (e *Engine) CanConfirm(actor string, count int) Decision {
	rule := e.ruleFor(actor)
	if !rule.CanConfirm {
		return decision(rule.Role, false, "actor cannot confirm selections")
	}
	if count < 1 {
		return decision(rule.Role, false, "at least one object is required")
	}
	if count > rule.MaximumSelection {
		return decision(rule.Role, false, "selection exceeds actor limit")
	}
	return decision(rule.Role, true, "confirmation allowed")
}

func (e *Engine) CanDispatch(actor string) Decision {
	rule := e.ruleFor(actor)
	return decision(rule.Role, rule.CanDispatch, "actor cannot dispatch records")
}

func (e *Engine) CanPublish(actor string) Decision {
	rule := e.ruleFor(actor)
	return decision(rule.Role, rule.CanPublish, "actor cannot publish records")
}

func decision(role Role, allowed bool, reason string) Decision {
	if allowed {
		return Decision{Allowed: true, Role: role, Reason: reason}
	}
	return Decision{Allowed: false, Role: role, Reason: reason}
}

func Require(decision Decision) error {
	if decision.Allowed {
		return nil
	}
	return fmt.Errorf("%w: %s", domain.ErrConflict, decision.Reason)
}

func Describe(engine *Engine, actor string) map[string]bool {
	return map[string]bool{"create": engine.CanCreate(actor).Allowed, "select": engine.CanSelect(actor, domain.ObjectTypePerson, 0).Allowed, "confirm": engine.CanConfirm(actor, 1).Allowed, "dispatch": engine.CanDispatch(actor).Allowed, "publish": engine.CanPublish(actor).Allowed}
}
