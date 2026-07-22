package journal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os/exec"

	"machine-observability/internal/collector"
)

type Collector struct {
	Logger        *slog.Logger
	parseFailures int
}

func (c *Collector) Name() string { return "journal" }

func (c *Collector) Run(ctx context.Context, events chan<- collector.Event) error {
	if c.Logger == nil {
		c.Logger = slog.New(slog.DiscardHandler)
	}

	cmd := exec.CommandContext(ctx, "journalctl", "-f", "-o", "json", "--no-pager")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	decoder := json.NewDecoder(stdout)
	var decodeErr error
	for {
		var raw map[string]any
		if err := decoder.Decode(&raw); err != nil {
			if !errors.Is(err, io.EOF) {
				decodeErr = err
			}
			break
		}

		event, err := Parse(raw)
		if err != nil {
			c.parseFailures++
			cursor, _ := raw["__CURSOR"].(string)
			c.Logger.Debug("parse failure", "error", err, "cursor", cursor)
			if c.parseFailures%100 == 0 {
				c.Logger.Warn("parse failures accumulating", "count", c.parseFailures)
			}
			continue
		}

		events <- event
	}

	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if decodeErr != nil {
		return decodeErr
	}
	return waitErr
}
