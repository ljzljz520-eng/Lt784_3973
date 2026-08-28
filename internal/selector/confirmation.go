package selector

import (
	"fmt"

	"example.com/permission-selector/internal/domain"
)

func (s *Service) Reject(requestID, actor, reason string) error {
	if requestID == "" || actor == "" || len(reason) < 3 {
		return domain.ErrInvalidInput
	}
	request, err := s.store.FindRequest(requestID)
	if err != nil {
		return err
	}
	if request.Status != domain.StatusConfirmed && request.Status != domain.StatusDispatched {
		return fmt.Errorf("%w: only confirmed requests can be rejected", domain.ErrConflict)
	}
	request.Status = domain.StatusRejected
	request.Version++
	request.UpdatedAt = s.clock()
	if err := s.store.SaveRequest(request); err != nil {
		return err
	}
	return s.audit.Write(requestID, actor, "reject", reason)
}

func (s *Service) Publish(requestID, actor string) (domain.AuthorizationRequest, error) {
	if requestID == "" || actor == "" {
		return domain.AuthorizationRequest{}, domain.ErrInvalidInput
	}
	request, err := s.store.FindRequest(requestID)
	if err != nil {
		return request, err
	}
	if err := domain.CanPublish(request.Status); err != nil {
		return request, err
	}
	request.Status = domain.StatusPublished
	request.Version++
	request.UpdatedAt = s.clock()
	if err := s.store.SaveRequest(request); err != nil {
		return request, err
	}
	if err := s.audit.Write(requestID, actor, "publish", "published confirmed authorization"); err != nil {
		return request, err
	}
	return request, nil
}

func (s *Service) Summary(requestID string) (domain.SelectionSummary, error) {
	objects, err := s.store.ListObjects(requestID)
	if err != nil {
		return domain.SelectionSummary{}, err
	}
	return domain.SummarizeObjects(objects), nil
}
