//go:build linux

package ebpfsource

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/example/capidata/badger/internal/domain"
)

// Loader consumes a pre-compiled capture.bpf.o. Compile it with clang/bpf2go in the image build.
// The loader intentionally keeps symbol attachment configurable because library paths differ by image.
type Loader struct {
	collection *ebpf.Collection
	reader     *ringbuf.Reader
	links      []link.Link
	once       sync.Once
}
type rawEvent struct {
	PIDTGID, CgroupID, Timestamp uint64
	Length                       uint32
	Direction, Source            uint8
	Payload                      [8192]byte
}

func New(object string) (*Loader, error) {
	spec, err := ebpf.LoadCollectionSpec(object)
	if err != nil {
		return nil, err
	}
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return nil, err
	}
	mapObj := coll.Maps["events"]
	if mapObj == nil {
		coll.Close()
		return nil, fmt.Errorf("events map missing")
	}
	r, err := ringbuf.NewReader(mapObj)
	if err != nil {
		coll.Close()
		return nil, err
	}
	return &Loader{collection: coll, reader: r}, nil
}

func (l *Loader) Attach(pid int, library, symbol string, programName string) error {
	prog := l.collection.Programs[programName]
	if prog == nil {
		return fmt.Errorf("program %q missing", programName)
	}
	lib, err := link.OpenExecutable(library)
	if err != nil {
		return err
	}
	hook, err := lib.Uprobe(symbol, prog, &link.UprobeOptions{PID: pid})
	if err != nil {
		return err
	}
	l.links = append(l.links, hook)
	return nil
}

func (l *Loader) Events(ctx context.Context) (<-chan domain.AuditEvent, error) {
	out := make(chan domain.AuditEvent)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			record, err := l.reader.Read()
			if err != nil {
				if err == ringbuf.ErrClosed {
					return
				}
				continue
			}
			if len(record.RawSample) < 32 {
				continue
			}
			length := binary.LittleEndian.Uint32(record.RawSample[24:28])
			if length > 8192 {
				length = 8192
			}
			event := domain.NewEvent(domain.HashValue(record.RawSample), fmt.Sprintf("%x", record.RawSample[8:16]), "", domain.Update, string(record.RawSample[32:32+length]), timeFromNS(binary.LittleEndian.Uint64(record.RawSample[16:24])))
			select {
			case out <- event:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func timeFromNS(ns uint64) (t time.Time) { return time.Unix(0, int64(ns)).UTC() }
func (l *Loader) Close() error {
	l.once.Do(func() {
		_ = l.reader.Close()
		for _, hook := range l.links {
			_ = hook.Close()
		}
		l.collection.Close()
	})
	return nil
}
