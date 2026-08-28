package workflow

import (
	"fmt"
	"strings"
	"time"

	"example.com/permission-selector/internal/audit"
	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/org"
	"example.com/permission-selector/internal/policy"
	"example.com/permission-selector/internal/selector"
	"example.com/permission-selector/internal/store"
)

type Service struct {
	store    *store.Store
	org      *org.Service
	selector *selector.Service
	audit    *audit.Service
	policy   *policy.Engine
	clock    func() time.Time
}

func NewService(database *store.Store, organization *org.Service, selection *selector.Service, auditLog *audit.Service) *Service {
	return &Service{store: database, org: organization, selector: selection, audit: auditLog, policy: policy.NewEngine(), clock: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) CreateRequest(command domain.RequestCommand) (domain.AuthorizationRequest, error) {
	if err := command.Validate(); err != nil {
		return domain.AuthorizationRequest{}, err
	}
	if err := policy.Require(s.policy.CanCreate(command.Actor)); err != nil {
		return domain.AuthorizationRequest{}, err
	}
	if _, err := s.store.FindRequest(command.RequestID); err == nil {
		return domain.AuthorizationRequest{}, fmt.Errorf("%w: request already exists", domain.ErrConflict)
	}
	now := s.clock()
	request := domain.AuthorizationRequest{ID: command.RequestID, Title: strings.TrimSpace(command.Title), Applicant: command.Actor, Description: strings.TrimSpace(command.Reason), Status: domain.StatusDraft, Version: 1, CreatedAt: now, UpdatedAt: now}
	if err := s.store.SaveRequest(request); err != nil {
		return domain.AuthorizationRequest{}, err
	}
	if err := s.audit.Write(request.ID, command.Actor, "create", request.Title); err != nil {
		return domain.AuthorizationRequest{}, err
	}
	return request, nil
}

func (s *Service) Accept(command domain.RequestCommand) (domain.WorkflowReceipt, error) {
	request, err := s.CreateRequest(command)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	object, err := s.selector.Select(domain.SelectionCommand{RequestID: request.ID, Actor: command.Actor, RefID: "root", Type: domain.ObjectTypeDepartment, Name: "全组织"})
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	_, summary, err := s.selector.Confirm(request.ID, command.Actor)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	dispatch, err := s.selector.Dispatch(request.ID, command.Actor, 1, 10)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	current, err := s.GetRequest(request.ID)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	count, err := s.audit.Count(request.ID)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	return domain.WorkflowReceipt{Request: current, Summary: summary, Dispatch: dispatch, SnapshotID: "snapshot-" + dispatch.ID, AuditEntries: count + boolCount(object.ID != "")}, nil
}

func boolCount(condition bool) int {
	if condition {
		return 1
	}
	return 0
}

func (s *Service) GetRequest(id string) (domain.AuthorizationRequest, error) {
	if strings.TrimSpace(id) == "" {
		return domain.AuthorizationRequest{}, domain.ErrInvalidInput
	}
	return s.store.FindRequest(id)
}

func (s *Service) Publish(requestID, actor string) (domain.AuthorizationRequest, error) {
	return s.selector.Publish(requestID, actor)
}

func (s *Service) ReopenAndRead(requestID string) (domain.ResultSnapshot, error) {
	if strings.TrimSpace(requestID) == "" {
		return domain.ResultSnapshot{}, domain.ErrInvalidInput
	}
	return s.selector.ResultSnapshot(requestID)
}
