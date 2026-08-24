package selector

import (
	"fmt"
	"time"

	"example.com/permission-selector/internal/audit"
	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/org"
	"example.com/permission-selector/internal/policy"
	"example.com/permission-selector/internal/store"
)

type Service struct {
	store  *store.Store
	org    *org.Service
	audit  *audit.Service
	policy *policy.Engine
	clock  func() time.Time
}

func NewService(database *store.Store, organization *org.Service, auditLog *audit.Service) *Service {
	return &Service{store: database, org: organization, audit: auditLog, policy: policy.NewEngine(), clock: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) LoadAccounts(filter domain.AccountFilter, page, pageSize int) (domain.Page[domain.Account], error) {
	if page < 1 || pageSize < 1 || pageSize > 100 {
		return domain.Page[domain.Account]{}, domain.ErrInvalidPage
	}
	accounts, err := s.org.AccountsForNode(filter)
	if err != nil {
		return domain.Page[domain.Account]{}, err
	}
	return domain.BuildAccountPage(accounts, page, pageSize)
}

func (s *Service) Select(command domain.SelectionCommand) (domain.AuthorizationObject, error) {
	if err := command.Validate(); err != nil {
		return domain.AuthorizationObject{}, err
	}
	request, err := s.store.FindRequest(command.RequestID)
	if err != nil {
		return domain.AuthorizationObject{}, err
	}
	if !domain.CanSelect(request.Status) {
		return domain.AuthorizationObject{}, fmt.Errorf("%w: request is %s", domain.ErrConflict, request.Status)
	}
	if command.Type == domain.ObjectTypePerson {
		account, findErr := s.store.FindAccount(command.RefID)
		if findErr != nil {
			return domain.AuthorizationObject{}, findErr
		}
		if !account.Active {
			return domain.AuthorizationObject{}, domain.ErrInactiveRecord
		}
	} else if _, findErr := s.store.FindNode(command.RefID); findErr != nil {
		return domain.AuthorizationObject{}, findErr
	}
	objects, err := s.store.ListObjects(command.RequestID)
	if err != nil {
		return domain.AuthorizationObject{}, err
	}
	for _, existing := range objects {
		if existing.RefID == command.RefID && existing.Type == command.Type {
			return domain.AuthorizationObject{}, domain.ErrDuplicate
		}
	}
	if err := policy.Require(s.policy.CanSelect(command.Actor, command.Type, len(objects))); err != nil {
		return domain.AuthorizationObject{}, err
	}
	object := domain.AuthorizationObject{ID: newID("object", command.RequestID, command.RefID), RequestID: command.RequestID, RefID: command.RefID, Type: command.Type, Name: command.Name, SelectedAt: s.clock()}
	if err := s.store.SaveObject(object); err != nil {
		return domain.AuthorizationObject{}, err
	}
	if err := s.audit.Write(command.RequestID, command.Actor, "select", fmt.Sprintf("selected %s %s", command.Type, command.RefID)); err != nil {
		return domain.AuthorizationObject{}, err
	}
	return object, nil
}

func (s *Service) Remove(requestID, actor, objectID string) error {
	if requestID == "" || actor == "" || objectID == "" {
		return domain.ErrInvalidInput
	}
	request, err := s.store.FindRequest(requestID)
	if err != nil {
		return err
	}
	if !domain.CanSelect(request.Status) {
		return domain.ErrConflict
	}
	object, err := s.store.FindObject(objectID)
	if err != nil {
		return err
	}
	if object.RequestID != requestID {
		return domain.ErrConflict
	}
	if err := s.store.DeleteObject(objectID); err != nil {
		return err
	}
	return s.audit.Write(requestID, actor, "remove", "removed "+objectID)
}

func (s *Service) CurrentSelection(requestID string) (domain.Page[domain.AuthorizationObject], error) {
	if requestID == "" {
		return domain.Page[domain.AuthorizationObject]{}, domain.ErrInvalidInput
	}
	objects, err := s.store.ListObjects(requestID)
	if err != nil {
		return domain.Page[domain.AuthorizationObject]{}, err
	}
	return domain.Page[domain.AuthorizationObject]{Items: objects, Page: 1, PageSize: len(objects), Total: len(objects)}, nil
}

func (s *Service) Confirm(requestID, actor string) (domain.AuthorizationRequest, domain.SelectionSummary, error) {
	request, err := s.store.FindRequest(requestID)
	if err != nil {
		return request, domain.SelectionSummary{}, err
	}
	objects, err := s.store.ListObjects(requestID)
	if err != nil {
		return request, domain.SelectionSummary{}, err
	}
	if err := domain.CanConfirm(request.Status, len(objects)); err != nil {
		return request, domain.SelectionSummary{}, err
	}
	request.Status = domain.StatusConfirmed
	request.Version++
	request.ConfirmedAt = s.clock()
	request.UpdatedAt = request.ConfirmedAt
	if err := s.store.SaveRequest(request); err != nil {
		return request, domain.SelectionSummary{}, err
	}
	if err := s.audit.Write(requestID, actor, "confirm", fmt.Sprintf("confirmed %d objects", len(objects))); err != nil {
		return request, domain.SelectionSummary{}, err
	}
	return request, domain.SummarizeObjects(objects), nil
}

func newID(prefix string, values ...string) string {
	result := prefix
	for _, value := range values {
		result += "-" + value
	}
	return result
}
