package cpu

import (
	"context"
	"fmt"
	"log/slog"
	"machine-observability/internal/collector"
	"os"
	"time"
)

const maxFailures = 5

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
				c.Logger.Warn("failed to sample", "error", err)
				if c.sampleFailures > maxFailures {
					return fmt.Errorf("sample failures exceeded threshold of %d, shutting down collector", maxFailures)
				}
				continue
			}
			c.sampleFailures = 0

			// Shutdown can interrupt this loop mid-tick, delivering only some of the tick's rows (the sink flushes whatever arrived).
			// Acceptable: each row is individually true; only tick-level aggregations (e.g. "all" vs sum of cores) won't balance for the final timestamp.
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
	defer func() { _ = f.Close() }()
	return parseStat(f, c.BootId, time.Now().UTC())
}
