package journal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"machine-observability/internal/collector"
)

type Collector struct {
	Logger        *slog.Logger
	CursorPath    string
	parseFailures int
}

func (c *Collector) Name() string { return "journal" }

func (c *Collector) consumeStream(ctx context.Context, r io.Reader, events chan<- collector.Event) error {
	decoder := json.NewDecoder(r)
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

		select {
		case events <- event:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return decodeErr
}

func (c *Collector) runOnce(ctx context.Context, events chan<- collector.Event, cursor string) error {
	args := []string{"-f", "-o", "json", "--no-pager"}

	if cursor != "" {
		args = append(args, "--after-cursor", cursor)
	}
	cmd := exec.CommandContext(ctx, "journalctl", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	decodeErr := c.consumeStream(ctx, stdout, events)

	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if decodeErr != nil {
		return decodeErr
	}
	return waitErr
}

func (c *Collector) Run(ctx context.Context, events chan<- collector.Event) error {
	if c.Logger == nil {
		c.Logger = slog.New(slog.DiscardHandler)
	}

	data, err := os.ReadFile(c.CursorPath)
	cursor := ""

	switch {
	case err == nil:
		cursor = strings.TrimSpace(string(data))
	case !errors.Is(err, fs.ErrNotExist):
		c.Logger.Warn("could not read cursor file, starting from now", "error", err)
	}

	delay := time.Second
	for {
		started := time.Now()
		err := c.runOnce(ctx, events, cursor)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if time.Since(started) > time.Minute {
			delay = time.Second
		}

		c.Logger.Warn("journalctl exited, restarting", "error", err, "uptime", time.Since(started), "backoff", delay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay = min(delay*2, 30*time.Second)
	}
}
