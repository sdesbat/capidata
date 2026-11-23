package usecase

import (
	"context"
	"time"

	"github.com/example/capidata/badger/internal/domain"
	"github.com/example/capidata/badger/internal/port"
)

type StatusService struct{ journal port.EventJournal }

func NewStatusService(journal port.EventJournal) *StatusService {
	return &StatusService{journal: journal}
}

func (s *StatusService) Record(ctx context.Context, event domain.AuditEvent) error {
	return s.journal.Append(ctx, event)
}

func (s *StatusService) Finalize(ctx context.Context, event domain.AuditEvent, status domain.Status, reason string) error {
	updated, err := event.WithStatus(status, reason, time.Now().UTC())
	if err != nil {
		return err
	}
	return s.journal.Append(ctx, updated)
}
