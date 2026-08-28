package workflow

import (
	"fmt"
	"time"

	"example.com/permission-selector/internal/domain"
)

func (s *Service) PublishWorkflow(requestID, actor string) (domain.WorkflowReceipt, error) {
	request, err := s.GetRequest(requestID)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	if err := domain.CanPublish(request.Status); err != nil {
		return domain.WorkflowReceipt{}, err
	}
	published, err := s.Publish(requestID, actor)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	summary, err := s.selector.Summary(requestID)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	dispatches, err := s.store.ListDispatches(requestID)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	if len(dispatches) == 0 {
		return domain.WorkflowReceipt{}, fmt.Errorf("%w: no dispatch record", domain.ErrConflict)
	}
	count, err := s.audit.Count(requestID)
	if err != nil {
		return domain.WorkflowReceipt{}, err
	}
	return domain.WorkflowReceipt{Request: published, Summary: summary, Dispatch: dispatches[len(dispatches)-1], SnapshotID: "snapshot-" + dispatches[len(dispatches)-1].ID, AuditEntries: count}, nil
}

func (s *Service) UpdateDescription(requestID, actor, description string) (domain.AuthorizationRequest, error) {
	if requestID == "" || actor == "" || len(description) < 5 {
		return domain.AuthorizationRequest{}, domain.ErrInvalidInput
	}
	request, err := s.store.FindRequest(requestID)
	if err != nil {
		return request, err
	}
	if request.Status != domain.StatusDraft && request.Status != domain.StatusRejected {
		return request, domain.ErrConflict
	}
	request.Description = description
	request.Version++
	request.UpdatedAt = timeNow()
	if err := s.store.SaveRequest(request); err != nil {
		return request, err
	}
	if err := s.audit.Write(requestID, actor, "update", "description updated"); err != nil {
		return request, err
	}
	return request, nil
}

func timeNow() (value time.Time) { return time.Now().UTC() }
