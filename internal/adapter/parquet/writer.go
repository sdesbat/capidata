package parquetwriter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/capidata/badger/internal/domain"
	"github.com/example/capidata/badger/internal/port"
	"github.com/parquet-go/parquet-go"
)

type Writer struct {
	Store  port.ObjectStore
	Prefix string
}

type row struct {
	EventID, TransactionID, StatementID, TenantID                         string
	Operation, Database, Schema, Table, SQLTemplate, Status, StatusReason string
	ObservedAt                                                            time.Time
	CommittedAt                                                           *time.Time
}

func New(store port.ObjectStore, prefix string) *Writer { return &Writer{Store: store, Prefix: prefix} }

func (w *Writer) Write(ctx context.Context, events []domain.AuditEvent) error {
	if len(events) == 0 {
		return nil
	}
	file, err := os.CreateTemp("", "badger-*.parquet")
	if err != nil {
		return err
	}
	name := file.Name()
	defer os.Remove(name)
	rows := make([]row, 0, len(events))
	for _, event := range events {
		if event.Status == domain.Committed {
			rows = append(rows, row{event.EventID, event.TransactionID, event.StatementID, event.TenantID, string(event.Operation), event.Database, event.Schema, event.Table, event.SQLTemplate, string(event.Status), event.StatusReason, event.ObservedAt, event.CommittedAt})
		}
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := parquet.WriteFile(name, rows); err != nil {
		return fmt.Errorf("write parquet: %w", err)
	}
	input, err := os.Open(name)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	key := filepath.Join(w.Prefix, fmt.Sprintf("event_date=%s", time.Now().UTC().Format("2006-01-02")), fmt.Sprintf("batch-%d.parquet", time.Now().UnixNano()))
	return w.Store.Put(ctx, key, input, info.Size())
}
