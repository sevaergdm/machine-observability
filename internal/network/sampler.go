package network

import (
	"context"
	"machine-observability/internal/collector"
	"os"
	"time"
)

type Sampler struct {
	BootId string
}

func (s *Sampler) Name() string { return "network" }

func (s *Sampler) Sample(ctx context.Context) ([]collector.Event, error) {
	var events []collector.Event

	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	entries, err := parseNetDev(f, s.BootId, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		events = append(events, entry)
	}
	return events, nil
}
