package journal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"

	"machine-observability/internal/collector"
	"machine-observability/internal/config"
)

type Collector struct {
	Config config.CollectorConfig
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
	for {
		var raw map[string]any

		err := decoder.Decode(&raw)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		event, err := Parse(raw)
		if err != nil {
			continue
		}

		events <- event
	}

	return cmd.Wait()
}
