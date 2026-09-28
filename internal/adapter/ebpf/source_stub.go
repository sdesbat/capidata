//go:build !linux

package ebpfsource

import (
	"context"
	"errors"
	"github.com/example/capidata/badger/internal/domain"
)

type Loader struct{}

func New(string) (*Loader, error) { return nil, errors.New("eBPF capture requires Linux") }
func (l *Loader) Events(context.Context) (<-chan domain.AuditEvent, error) {
	return nil, errors.New("eBPF capture requires Linux")
}
func (l *Loader) Close() error { return nil }
