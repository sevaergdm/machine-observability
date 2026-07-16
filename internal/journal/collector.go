package journal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
)

type Collector struct {
}

func (c Collector) Run(ctx context.Context, events chan<- JournalEntry) error {
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
