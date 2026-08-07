package cpu

import (
	"context"
	"fmt"
	"log/slog"
	"machine-observability/internal/collector"
	"os"
	"time"
)

const MAX_FAILURES = 5

type Collector struct {
	Logger         *slog.Logger
	BootId         string
	Interval       time.Duration
	sampleFailures int
}

func (c *Collector) Name() string { return "cpu" }

func (c *Collector) Run(ctx context.Context, events chan<- collector.Event) error {
	if c.Logger == nil {
		c.Logger = slog.New(slog.DiscardHandler)
	}

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			entries, err := c.sample()
			if err != nil {
				c.sampleFailures++
				c.Logger.Debug("failed to sample", "error", err)
				if c.sampleFailures > MAX_FAILURES {
					return fmt.Errorf("sample failures exceeded threshold of %d, shutting down collector", MAX_FAILURES)
				}
			}
			c.sampleFailures = 0

			for _, entry := range entries {
				select {
				case events <- entry:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

		case <-ctx.Done():
			return ctx.Err()

		}
	}
}

func (c *Collector) sample() ([]Entry, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseStat(f, c.BootId, time.Now().UTC())
}
