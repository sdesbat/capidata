package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Status string

const (
	Pending   Status = "pending"
	Committed Status = "committed"
	Failed    Status = "failed"
	Unknown   Status = "unknown"
)

type Operation string

const (
	Insert Operation = "INSERT"
	Update Operation = "UPDATE"
)

type Parameter struct {
	Position int    `json:"position" parquet:"position"`
	Type     string `json:"type,omitempty" parquet:"type"`
	Value    string `json:"value,omitempty" parquet:"value"`
	Hash     string `json:"hash,omitempty" parquet:"hash"`
	Format   string `json:"format,omitempty" parquet:"format"`
	Redacted bool   `json:"redacted" parquet:"redacted"`
}

type Identity struct {
	Namespace string `json:"namespace" parquet:"namespace"`
	Pod       string `json:"pod,omitempty" parquet:"pod"`
	Container string `json:"container,omitempty" parquet:"container"`
	Node      string `json:"node,omitempty" parquet:"node"`
}

type AuditEvent struct {
	EventID       string      `json:"event_id" parquet:"event_id"`
	TransactionID string      `json:"transaction_id" parquet:"transaction_id"`
	StatementID   string      `json:"statement_id" parquet:"statement_id"`
	TenantID      string      `json:"tenant_id" parquet:"tenant_id"`
	Database      string      `json:"database,omitempty" parquet:"database"`
	Schema        string      `json:"schema,omitempty" parquet:"schema"`
	Table         string      `json:"table,omitempty" parquet:"table"`
	Operation     Operation   `json:"operation" parquet:"operation"`
	SQLTemplate   string      `json:"sql_template" parquet:"sql_template"`
	Parameters    []Parameter `json:"parameters,omitempty" parquet:"parameters"`
	Status        Status      `json:"status" parquet:"status"`
	StatusReason  string      `json:"status_reason,omitempty" parquet:"status_reason"`
	ObservedAt    time.Time   `json:"observed_at" parquet:"observed_at"`
	CommittedAt   *time.Time  `json:"committed_at,omitempty" parquet:"committed_at"`
	Source        Identity    `json:"source" parquet:"source"`
}

func NewEvent(id, tx, tenant string, op Operation, sql string, now time.Time) AuditEvent {
	return AuditEvent{EventID: id, TransactionID: tx, StatementID: id, TenantID: tenant, Operation: op, SQLTemplate: sql, Status: Pending, ObservedAt: now}
}

func (e AuditEvent) WithStatus(status Status, reason string, at time.Time) (AuditEvent, error) {
	if e.Status != Pending && e.Status != Unknown {
		return e, fmt.Errorf("event %s already finalized as %s", e.EventID, e.Status)
	}
	if status != Committed && status != Failed && status != Unknown {
		return e, errors.New("final status must be committed, failed, or unknown")
	}
	e.Status, e.StatusReason = status, reason
	if status == Committed {
		e.CommittedAt = &at
	}
	return e, nil
}

func HashValue(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func ParseOperation(sql string) (Operation, bool) {
	fields := strings.Fields(strings.ToUpper(sql))
	if len(fields) == 0 {
		return "", false
	}
	return Operation(fields[0]), fields[0] == string(Insert) || fields[0] == string(Update)
}
