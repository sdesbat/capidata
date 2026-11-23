package journal

import (
	"context"
	"testing"
	"time"

	"github.com/example/capidata/badger/internal/domain"
)

func TestJSONLCommittedProjection(t *testing.T) {
	j, err := OpenJSONL(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := domain.NewEvent("e1", "tx", "ns", domain.Insert, "INSERT INTO t VALUES ($1)", time.Now())
	if err = j.Append(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	e, err = e.WithStatus(domain.Committed, "commit_observed", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err = j.Append(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	items, err := j.ReadCommitted(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Status != domain.Committed {
		t.Fatalf("unexpected items: %+v", items)
	}
}
