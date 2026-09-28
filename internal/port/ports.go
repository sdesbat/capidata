package port

import (
	"context"
	"io"

	"github.com/example/capidata/badger/internal/domain"
)

type EventJournal interface {
	Append(context.Context, domain.AuditEvent) error
	ReadCommitted(context.Context) ([]domain.AuditEvent, error)
	Seal(context.Context) error
}

type EventSource interface {
	Events(context.Context) (<-chan domain.AuditEvent, error)
	io.Closer
}

type ObjectStore interface {
	Put(context.Context, string, io.Reader, int64) error
}
