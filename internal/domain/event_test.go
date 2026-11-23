package domain

import (
	"testing"
	"time"
)

func TestEventStatusTransitions(t *testing.T) {
	event := NewEvent("e1", "tx1", "ns", Update, "UPDATE t SET x=$1", time.Now())
	committed, err := event.WithStatus(Committed, "commit_observed", time.Now())
	if err != nil || committed.Status != Committed || committed.CommittedAt == nil {
		t.Fatalf("commit transition failed: %+v %v", committed, err)
	}
	if _, err := committed.WithStatus(Failed, "late_error", time.Now()); err == nil {
		t.Fatal("expected finalized event rejection")
	}
}
