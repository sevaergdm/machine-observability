package sink

import (
	"context"
	"machine-observability/internal/collector"
	"testing"
	"time"
)

type fakeEvent struct{ n int }

func (f fakeEvent) Source() string       { return "fake" }
func (f fakeEvent) Timestamp() time.Time { return time.Time{} }

func TestFlushOnRowCount(t *testing.T) {
	events := make(chan collector.Event)
	flushes := make(chan []collector.Event, 4)

	w := NewWriter(events, func(rows []collector.Event) error {
		flushes <- rows
		return nil
	}, 3, time.Hour, nil)

	go w.Run(context.Background())

	for i := range 3 {
		events <- fakeEvent{n: i}
	}

	select {
	case batch := <-flushes:
		if len(batch) != 3 {
			t.Errorf("flushed %d rows, want 3", len(batch))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no flush happened")
	}
}

func TestFlushOnAge(t *testing.T) {
	events := make(chan collector.Event)
	flushes := make(chan []collector.Event, 1)

	w := NewWriter(events, func(rows []collector.Event) error {
		flushes <- rows
		return nil
	}, 1000, 50*time.Millisecond, nil)

	go w.Run(context.Background())

	events <- fakeEvent{n: 1}

	select {
	case batch := <-flushes:
		if len(batch) != 1 {
			t.Errorf("flushed %d rows, want 1", len(batch))
		}
	case <-time.After(3 * time.Second):
		t.Errorf("did not flush after 3 seconds")
	}
}

func TestFlushOnCancel(t *testing.T) {
	events := make(chan collector.Event)
	flushes := make(chan []collector.Event, 2)

	w := NewWriter(events, func(rows []collector.Event) error {
		flushes <- rows
		return nil
	}, 2, time.Second, nil)

	done := make(chan error, 1)
	go func() { done <- w.Run(context.Background()) }()

	events <- fakeEvent{n: 1}
	events <- fakeEvent{n: 2}
	close(events)

	if err := <-done; err != nil {
		t.Errorf("expected no error, but got: %v", err)
	}

	select {
	case batch := <-flushes:
		if len(batch) != 2 {
			t.Errorf("flushed %d rows, want 2", len(batch))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no flush happened")
	}
}
