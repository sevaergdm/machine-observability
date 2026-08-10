package memory

import (
	"context"
	"machine-observability/internal/collector"
	"os"
	"time"
)

type Sampler struct {
	BootId string
}

func (s *Sampler) Name() string { return "memory" }

func (s *Sampler) Sample(ctx context.Context) ([]collector.Event, error) {
	var events []collector.Event
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	event, err := parseMemInfo(f, s.BootId, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	events = append(events, event)
	return events, nil
}
