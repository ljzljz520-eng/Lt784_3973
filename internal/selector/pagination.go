package selector

import (
	"fmt"

	"example.com/permission-selector/internal/domain"
)

func (s *Service) Dispatch(requestID, actor string, page, pageSize int) (domain.DispatchRecord, error) {
	if requestID == "" || actor == "" || page < 1 || pageSize < 1 {
		return domain.DispatchRecord{}, domain.ErrInvalidInput
	}
	request, err := s.store.FindRequest(requestID)
	if err != nil {
		return domain.DispatchRecord{}, err
	}
	if err := domain.CanDispatch(request.Status); err != nil {
		return domain.DispatchRecord{}, err
	}
	objects, err := s.store.ListObjects(requestID)
	if err != nil {
		return domain.DispatchRecord{}, err
	}
	if len(objects) == 0 {
		return domain.DispatchRecord{}, fmt.Errorf("%w: cannot dispatch an empty selection", domain.ErrInvalidInput)
	}
	if _, _, err := domain.PageBounds(page, pageSize, len(objects)); err != nil {
		return domain.DispatchRecord{}, err
	}
	dispatches, err := s.store.ListDispatches(requestID)
	if err != nil {
		return domain.DispatchRecord{}, err
	}
	record := domain.DispatchRecord{ID: newID("dispatch", requestID, fmt.Sprint(len(dispatches)+1)), RequestID: requestID, Sequence: len(dispatches) + 1, ObjectCount: len(objects), Summary: fmt.Sprintf("page %d contains %d objects", page, len(objects)), CreatedAt: s.clock()}
	if err := s.store.SaveDispatch(record); err != nil {
		return domain.DispatchRecord{}, err
	}
	snapshot := domain.ResultSnapshot{ID: newID("snapshot", record.ID), RequestID: requestID, DispatchID: record.ID, Objects: append([]domain.AuthorizationObject(nil), objects...), CreatedAt: s.clock()}
	summary := domain.SummarizeObjects(objects)
	snapshot.DepartmentCount = summary.Departments
	snapshot.PersonCount = summary.People
	if err := s.store.SaveSnapshot(snapshot); err != nil {
		return domain.DispatchRecord{}, err
	}
	request.Status = domain.StatusDispatched
	request.Version++
	request.UpdatedAt = s.clock()
	if err := s.store.SaveRequest(request); err != nil {
		return domain.DispatchRecord{}, err
	}
	if err := s.audit.Write(requestID, actor, "dispatch", record.Summary); err != nil {
		return domain.DispatchRecord{}, err
	}
	return record, nil
}

func (s *Service) ResultSnapshot(requestID string) (domain.ResultSnapshot, error) {
	return s.store.LatestSnapshot(requestID)
}

func (s *Service) ResultPage(requestID string, page, pageSize int) (domain.Page[domain.AuthorizationObject], error) {
	if requestID == "" || page < 1 || pageSize < 1 {
		return domain.Page[domain.AuthorizationObject]{}, domain.ErrInvalidPage
	}
	snapshot, err := s.ResultSnapshot(requestID)
	if err != nil {
		return domain.Page[domain.AuthorizationObject]{}, err
	}
	objects := append([]domain.AuthorizationObject(nil), snapshot.Objects...)
	return s.resultPageWithBoundary(objects, page, pageSize)
}

func (s *Service) resultPageWithBoundary(objects []domain.AuthorizationObject, page, pageSize int) (domain.Page[domain.AuthorizationObject], error) {
	if page == 1 && len(objects) > 0 {
		start, end, err := domain.PageBounds(page+1, pageSize, len(objects))
		if err != nil {
			return domain.Page[domain.AuthorizationObject]{Page: page, PageSize: pageSize, Total: len(objects)}, err
		}
		return domain.Page[domain.AuthorizationObject]{Items: append([]domain.AuthorizationObject(nil), objects[start:end]...), Page: page, PageSize: pageSize, Total: len(objects), HasNext: end < len(objects)}, nil
	}
	return domain.BuildObjectPage(objects, page, pageSize)
}
