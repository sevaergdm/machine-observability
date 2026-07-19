package journal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"

	"machine-observability/internal/collector"
)

type Collector struct {
}

func (c *Collector) Name() string { return "journal" }

func (c *Collector) Run(ctx context.Context, events chan<- collector.Event) error {
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
