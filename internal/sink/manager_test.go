package sink

import (
	"log/slog"
	"machine-observability/internal/collector"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var gotA, gotB [][]collector.Event

var flushA = func(rows []collector.Event) error { gotA = append(gotA, rows); return nil }
var flushB = func(rows []collector.Event) error { gotB = append(gotB, rows); return nil }

var tuning = Tuning{MaxRows: 100, MaxAge: time.Hour}

func TestManagerRoutesBySource(t *testing.T) {
	events := make(chan collector.Event)
	m := NewManager(events, slog.New(slog.DiscardHandler))
	if err := m.Register("a", flushA, tuning); err != nil {
		t.Fatalf("unexpected error registering source 'a': %v", err)
	}
	if err := m.Register("b", flushB, tuning); err != nil {
		t.Fatalf("unexpected error registering source 'b': %v", err)
	}

	go func() {
		events <- fakeEvent{Src: "a", N: 1}
		events <- fakeEvent{Src: "b", N: 2}
		events <- fakeEvent{Src: "ghost", N: 3}
		events <- fakeEvent{Src: "a", N: 4}
		close(events)
	}()

	m.Run()

	want := [][]collector.Event{{fakeEvent{Src: "a", N: 1}, fakeEvent{Src: "a", N: 4}}}
	if diff := cmp.Diff(want, gotA); diff != "" {
		t.Errorf("source a batches (-want, +got):\n%s", diff)
	}

	want = [][]collector.Event{{fakeEvent{Src: "b", N: 2}}}
	if diff := cmp.Diff(want, gotB); diff != "" {
		t.Errorf("source b batches (-want, +got):\n%s", diff)
	}
}

func TestManagerDuplicateRoute(t *testing.T) {
	events := make(chan collector.Event)
	m := NewManager(events, slog.New(slog.DiscardHandler))
	if err := m.Register("a", flushA, tuning); err != nil {
		t.Fatalf("unexpected error registering source 'a': %v", err)
	}

	if err := m.Register("a", flushA, tuning); err == nil {
		t.Errorf("expected duplicate error, but got none")
	}
}

func TestManagerLateStart(t *testing.T) {
	events := make(chan collector.Event)
	m := NewManager(events, slog.New(slog.DiscardHandler))
	close(events)
	m.Run()

	if err := m.Register("a", flushA, tuning); err == nil {
		t.Errorf("expected late start error, but got none")
	}

}
