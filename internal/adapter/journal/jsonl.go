package journal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/capidata/badger/internal/domain"
)

type JSONL struct {
	root string
	file *os.File
}

func OpenJSONL(root string) (*JSONL, error) {
	if err := os.MkdirAll(root, 0750); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(root, "events.open.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	return &JSONL{root: root, file: f}, nil
}

func (j *JSONL) Append(_ context.Context, event domain.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(j.file, "%s\n", data); err != nil {
		return err
	}
	return j.file.Sync()
}

func (j *JSONL) ReadCommitted(_ context.Context) ([]domain.AuditEvent, error) {
	entries := make(map[string]domain.AuditEvent)
	err := filepath.WalkDir(j.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return err
		}
		f, e := os.Open(path)
		if e != nil {
			return e
		}
		defer f.Close()
		s := bufio.NewScanner(f)
		s.Buffer(make([]byte, 64*1024), 4*1024*1024)
		for s.Scan() {
			var event domain.AuditEvent
			if e := json.Unmarshal(s.Bytes(), &event); e != nil {
				return e
			}
			entries[event.EventID] = event
		}
		return s.Err()
	})
	if err != nil {
		return nil, err
	}
	result := make([]domain.AuditEvent, 0, len(entries))
	for _, event := range entries {
		if event.Status == domain.Committed {
			result = append(result, event)
		}
	}
	return result, nil
}

func (j *JSONL) Seal(_ context.Context) error {
	if err := j.file.Sync(); err != nil {
		return err
	}
	if err := j.file.Close(); err != nil {
		return err
	}
	return os.Rename(filepath.Join(j.root, "events.open.jsonl"), filepath.Join(j.root, "events.ready.jsonl"))
}

func (j *JSONL) Close() error { return j.file.Close() }
