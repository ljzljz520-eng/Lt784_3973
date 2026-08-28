package store

import (
	"sort"

	"example.com/permission-selector/internal/domain"
)

func (s *Store) SaveAudit(event domain.AuditEvent) error {
	if event.ID == "" || event.RequestID == "" || event.Action == "" {
		return domain.ErrInvalidInput
	}
	return s.put(BucketAudits, event.ID, event)
}

func (s *Store) ListAudits(requestID string) ([]domain.AuditEvent, error) {
	values, err := listValues(s, BucketAudits, func(data []byte) (domain.AuditEvent, error) {
		var event domain.AuditEvent
		if err := decode(data, &event); err != nil {
			return event, err
		}
		return event, nil
	})
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.AuditEvent, 0, len(values))
	for _, event := range values {
		if event.RequestID == requestID {
			filtered = append(filtered, event)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.Before(filtered[j].CreatedAt) })
	return filtered, nil
}

func (s *Store) SaveDispatch(record domain.DispatchRecord) error {
	if record.ID == "" || record.RequestID == "" || record.Sequence < 1 {
		return domain.ErrInvalidInput
	}
	return s.put(BucketDispatches, record.ID, record)
}

func (s *Store) FindDispatch(id string) (domain.DispatchRecord, error) {
	var record domain.DispatchRecord
	err := s.get(BucketDispatches, id, &record)
	return record, err
}

func (s *Store) ListDispatches(requestID string) ([]domain.DispatchRecord, error) {
	values, err := listValues(s, BucketDispatches, func(data []byte) (domain.DispatchRecord, error) {
		var record domain.DispatchRecord
		if err := decode(data, &record); err != nil {
			return record, err
		}
		return record, nil
	})
	if err != nil {
		return nil, err
	}
	filtered := values[:0]
	for _, record := range values {
		if record.RequestID == requestID {
			filtered = append(filtered, record)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Sequence < filtered[j].Sequence })
	return filtered, nil
}

func (s *Store) SaveSnapshot(snapshot domain.ResultSnapshot) error {
	if snapshot.ID == "" || snapshot.RequestID == "" || snapshot.DispatchID == "" {
		return domain.ErrInvalidInput
	}
	return s.put(BucketSnapshots, snapshot.ID, snapshot)
}

func (s *Store) FindSnapshot(id string) (domain.ResultSnapshot, error) {
	var snapshot domain.ResultSnapshot
	err := s.get(BucketSnapshots, id, &snapshot)
	return snapshot, err
}

func (s *Store) ListSnapshots(requestID string) ([]domain.ResultSnapshot, error) {
	values, err := listValues(s, BucketSnapshots, func(data []byte) (domain.ResultSnapshot, error) {
		var snapshot domain.ResultSnapshot
		if err := decode(data, &snapshot); err != nil {
			return snapshot, err
		}
		return snapshot, nil
	})
	if err != nil {
		return nil, err
	}
	filtered := values[:0]
	for _, snapshot := range values {
		if snapshot.RequestID == requestID {
			filtered = append(filtered, snapshot)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.Before(filtered[j].CreatedAt) })
	return filtered, nil
}
