package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"example.com/permission-selector/internal/domain"
	bolt "go.etcd.io/bbolt"
)

const (
	BucketNodes      = "organization_nodes"
	BucketAccounts   = "accounts"
	BucketRequests   = "authorization_requests"
	BucketObjects    = "authorization_objects"
	BucketAudits     = "audit_events"
	BucketDispatches = "dispatch_records"
	BucketSnapshots  = "result_snapshots"
)

var allBuckets = [][]byte{
	[]byte(BucketNodes), []byte(BucketAccounts), []byte(BucketRequests),
	[]byte(BucketObjects), []byte(BucketAudits), []byte(BucketDispatches), []byte(BucketSnapshots),
}

type Store struct {
	db   *bolt.DB
	path string
	mu   sync.RWMutex
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("%w: database path is required", domain.ErrInvalidInput)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second, NoSync: false})
	if err != nil {
		return nil, fmt.Errorf("open bbolt database: %w", err)
	}
	store := &Store{db: db, path: path}
	if err := store.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range allBuckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string { return s.path }

func encode(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode record: %w", err)
	}
	return encoded, nil
}

func decode(data []byte, target any) error {
	if len(data) == 0 {
		return domain.ErrNotFound
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode record: %w", err)
	}
	return nil
}

func (s *Store) put(bucket, key string, value any) error {
	if key == "" {
		return fmt.Errorf("%w: empty key", domain.ErrInvalidInput)
	}
	data, err := encode(value)
	if err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s is missing", bucket)
		}
		return b.Put([]byte(key), data)
	})
}

func (s *Store) get(bucket, key string, target any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s is missing", bucket)
		}
		return decode(b.Get([]byte(key)), target)
	})
}

func (s *Store) remove(bucket, key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s is missing", bucket)
		}
		return b.Delete([]byte(key))
	})
}

func listValues[T any](s *Store, bucket string, decodeValue func([]byte) (T, error)) ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	values := make([]T, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %s is missing", bucket)
		}
		return b.ForEach(func(_, value []byte) error {
			item, err := decodeValue(value)
			if err != nil {
				return err
			}
			values = append(values, item)
			return nil
		})
	})
	return values, err
}
