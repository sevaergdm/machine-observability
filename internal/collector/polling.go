package collector

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type polling struct {
	sampler  Sampler
	interval time.Duration
	logger   *slog.Logger
	failures int
}

func (p *polling) Name() string { return p.sampler.Name() }
func (p *polling) Run(ctx context.Context, events chan<- Event) error {
	if p.logger == nil {
		p.logger = slog.New(slog.DiscardHandler)
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			entries, err := p.sampler.Sample(ctx)
			if err != nil {
				p.failures++
				p.logger.Warn("failed to sample", "error", err, "source", p.Name())
				if p.failures > maxFailures {
					return fmt.Errorf("sample failures exceeded threshold of %d, shutting down the collector: %s", maxFailures, p.Name())
				}
				continue
			}
			p.failures = 0

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

func NewPolling(s Sampler, interval time.Duration, logger *slog.Logger) Collector {
	return &polling{
		sampler:  s,
		interval: interval,
		logger:   logger,
	}
}
