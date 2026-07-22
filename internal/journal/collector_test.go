package journal

import (
	"context"
	"log/slog"
	"machine-observability/internal/collector"
	"strings"
	"testing"
)

const validLine = `{"__CURSOR":"c1","__REALTIME_TIMESTAMP":"1753142400000000","__MONOTONIC_TIMESTAMP":"1","__SEQNUM":"1","__SEQNUM_ID":"s1"}`
const noCursorLine = `{"__REALTIME_TIMESTAMP":"1753142400000000","__MONOTONIC_TIMESTAMP":"1","__SEQNUM":"1","__SEQNUM_ID":"s1"}`
const garbageLine = `{this is not json`

func TestConsumeStreamError(t *testing.T) {
	input := strings.Join([]string{validLine, validLine, noCursorLine, validLine, garbageLine}, "\n")

	c := &Collector{Logger: slog.New(slog.DiscardHandler)}

	events := make(chan collector.Event, 16)

	err := c.consumeStream(context.Background(), strings.NewReader(input), events)
	if err == nil {
		t.Error("expected an error from the garbage line, got nil")
	}

	if got := len(events); got != 3 {
		t.Errorf("events = %d, want 3", got)
	}

	if c.parseFailures != 1 {
		t.Errorf("parseFailures = %d, want 1", c.parseFailures)
	}
}

func TestConsumeStreamValid(t *testing.T) {
	input := validLine

	c := &Collector{Logger: slog.New(slog.DiscardHandler)}

	events := make(chan collector.Event, 16)

	err := c.consumeStream(context.Background(), strings.NewReader(input), events)
	if err != nil {
		t.Errorf("expected no error, but got %v", err)
	}

	if got := len(events); got != 1 {
		t.Errorf("events = %d, want 1", got)
	}

	if c.parseFailures != 0 {
		t.Errorf("parseFailures = %d, want 0", c.parseFailures)
	}
}
