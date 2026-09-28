package usecase

import (
	"strings"
	"time"

	"github.com/example/capidata/badger/internal/adapter/postgresprotocol"
	"github.com/example/capidata/badger/internal/domain"
)

type Correlator struct {
	Tenant   string
	Now      func() time.Time
	prepared map[string]string
	pending  []*domain.AuditEvent
}

func NewCorrelator(tenant string) *Correlator {
	return &Correlator{Tenant: tenant, Now: func() time.Time { return time.Now().UTC() }, prepared: map[string]string{}}
}

func (c *Correlator) Consume(m postgresprotocol.Message) []domain.AuditEvent {
	now := c.Now()
	var output []domain.AuditEvent
	switch m.Type {
	case 'P':
		if parsed, err := postgresprotocol.ParseParse(m); err == nil {
			c.prepared[parsed.Name] = parsed.Query
		}
	case 'B':
		if parsed, err := postgresprotocol.ParseBind(m); err == nil {
			if sql := c.prepared[parsed.Statement]; sql != "" {
				if op, ok := domain.ParseOperation(sql); ok {
					event := domain.NewEvent(domain.HashValue([]byte(sql+now.String())), "", c.Tenant, op, sql, now)
					for i, value := range parsed.Parameters {
						event.Parameters = append(event.Parameters, domain.Parameter{Position: i + 1, Value: string(value), Format: "text"})
					}
					c.pending = append(c.pending, &event)
				}
			}
		}
	case 'Q':
		if sql, err := postgresprotocol.CString(m.Body, new(int)); err == nil {
			if op, ok := domain.ParseOperation(sql); ok {
				event := domain.NewEvent(domain.HashValue([]byte(sql+now.String())), "", c.Tenant, op, sql, now)
				c.pending = append(c.pending, &event)
			}
		}
	case 'C':
		tag, _ := postgresprotocol.CommandTag(m)
		upper := strings.ToUpper(tag)
		if strings.HasPrefix(upper, "INSERT") || strings.HasPrefix(upper, "UPDATE") {
			output = append(output, c.finalize(domain.Committed, "command_complete", now)...)
		}
	case 'E':
		output = append(output, c.finalize(domain.Failed, "server_error", now)...)
	case 'Z':
		if len(m.Body) == 1 && m.Body[0] == 'E' {
			output = append(output, c.finalize(domain.Unknown, "failed_transaction_state", now)...)
		}
	}
	return output
}

func (c *Correlator) finalize(status domain.Status, reason string, now time.Time) []domain.AuditEvent {
	result := make([]domain.AuditEvent, 0, len(c.pending))
	for _, event := range c.pending {
		updated, err := event.WithStatus(status, reason, now)
		if err == nil {
			result = append(result, updated)
		}
	}
	c.pending = nil
	return result
}
