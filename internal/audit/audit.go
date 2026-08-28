package audit

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"example.com/permission-selector/internal/domain"
	"example.com/permission-selector/internal/store"
)

type Service struct {
	store *store.Store
	clock func() time.Time
	seq   uint64
}

func NewService(database *store.Store) *Service {
	return &Service{store: database, clock: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Write(requestID, actor, action, detail string) error {
	if strings.TrimSpace(requestID) == "" || strings.TrimSpace(actor) == "" || strings.TrimSpace(action) == "" {
		return domain.ErrInvalidInput
	}
	if len(detail) > 500 {
		return fmt.Errorf("%w: audit detail is too long", domain.ErrInvalidInput)
	}
	event := domain.AuditEvent{ID: fmt.Sprintf("audit-%d", atomic.AddUint64(&s.seq, 1)), RequestID: requestID, Actor: actor, Action: action, Detail: detail, CreatedAt: s.clock()}
	return s.store.SaveAudit(event)
}

func (s *Service) History(requestID string) ([]domain.AuditEvent, error) {
	if requestID == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.store.ListAudits(requestID)
}

func (s *Service) Count(requestID string) (int, error) {
	history, err := s.History(requestID)
	if err != nil {
		return 0, err
	}
	return len(history), nil
}
