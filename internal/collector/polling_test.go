package collector

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

type fakeEvent struct{ N int }

func (f fakeEvent) Source() string       { return "fake" }
func (f fakeEvent) Timestamp() time.Time { return time.Time{} }

type fakeSampler struct {
	fn func(ctx context.Context) ([]Event, error)
}

func (f *fakeSampler) Name() string                                { return "fake" }
func (f *fakeSampler) Sample(ctx context.Context) ([]Event, error) { return f.fn(ctx) }

func TestPollingEmitsOnTick(t *testing.T) {
	s := &fakeSampler{
		fn: func(ctx context.Context) ([]Event, error) {
			return []Event{fakeEvent{N: 1}, fakeEvent{N: 2}}, nil
		},
	}

	events := make(chan Event, 100)
	ctx, cancel := context.WithCancel(t.Context())

	p := NewPolling(s, 10*time.Millisecond, slog.New(slog.DiscardHandler))

	done := make(chan error, 1)
	go func() {
		done <- p.Run(ctx, events)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}

	if got := len(events); got < 2 {
		t.Errorf("got %d events, expected at least 2", got)
	}
}

func TestPollingFailAndRecover(t *testing.T) {
	calls := 0
	s := &fakeSampler{
		fn: func(ctx context.Context) ([]Event, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("boom")
			}
			return []Event{fakeEvent{N: calls}}, nil
		},
	}

	events := make(chan Event, 100)
	ctx, cancel := context.WithCancel(t.Context())

	p := NewPolling(s, 10*time.Millisecond, slog.New(slog.DiscardHandler))

	done := make(chan error, 1)
	go func() {
		done <- p.Run(ctx, events)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}

	if got := len(events); got < 2 {
		t.Errorf("got %d events, expected at least 2", got)
	}
}

func TestPollingErrorReturn(t *testing.T) {
	s := &fakeSampler{
		fn: func(ctx context.Context) ([]Event, error) {
			return nil, errors.New("some error")
		},
	}

	events := make(chan Event, 100)
	p := NewPolling(s, 10*time.Millisecond, slog.New(slog.DiscardHandler))

	done := make(chan error, 1)
	go func() {
		done <- p.Run(t.Context(), events)
	}()

	select {
	case err := <-done:
		if err == nil || errors.Is(err, context.Canceled) {
			t.Errorf("unexpected error text: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not self disable")
	}

	if got := len(events); got > 0 {
		t.Errorf("got %d events, expected 0", got)
	}
}

func TestPollingConsecutiveNotCumulative(t *testing.T) {
	calls := 0
	s := &fakeSampler{
		fn: func(ctx context.Context) ([]Event, error) {
			calls++
			if calls%2 == 1 {
				return nil, errors.New("boom")
			}
			return []Event{fakeEvent{N: calls}}, nil
		},
	}

	interval := 10 * time.Millisecond
	ticks := 3 * maxFailures

	events := make(chan Event, 100)
	p := NewPolling(s, interval, slog.New(slog.DiscardHandler))

	done := make(chan error, 1)
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		done <- p.Run(ctx, events)
	}()

	time.Sleep(time.Duration(ticks) * interval)

	select {
	case err := <-done:
		t.Fatalf("runner exited during alternating failures: %v", err)
	default:
	}

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}
