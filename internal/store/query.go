package store

import (
	"fmt"
	"sort"

	"example.com/permission-selector/internal/domain"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) Count(bucket string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return 0, fmt.Errorf("store is closed")
	}
	count := 0
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s is missing", bucket)
		}
		return b.ForEach(func(_, _ []byte) error { count++; return nil })
	})
	return count, err
}

func (s *Store) ActiveAccounts() ([]domain.Account, error) {
	accounts, err := s.ListAccounts()
	if err != nil {
		return nil, err
	}
	active := accounts[:0]
	for _, account := range accounts {
		if account.Active {
			active = append(active, account)
		}
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Display < active[j].Display })
	return active, nil
}

func (s *Store) LatestSnapshot(requestID string) (domain.ResultSnapshot, error) {
	snapshots, err := s.ListSnapshots(requestID)
	if err != nil {
		return domain.ResultSnapshot{}, err
	}
	if len(snapshots) == 0 {
		return domain.ResultSnapshot{}, domain.ErrNotFound
	}
	sort.Slice(snapshots, func(i, j int) bool { return snapshots[i].CreatedAt.Before(snapshots[j].CreatedAt) })
	return snapshots[len(snapshots)-1], nil
}

func (s *Store) CheckRequestReferences(requestID string) error {
	request, err := s.FindRequest(requestID)
	if err != nil {
		return err
	}
	objects, err := s.ListObjects(requestID)
	if err != nil {
		return err
	}
	for _, object := range objects {
		if object.RequestID != request.ID || object.RefID == "" {
			return domain.ErrConflict
		}
	}
	return nil
}
