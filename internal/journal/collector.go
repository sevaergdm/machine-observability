package journal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	lastCursor    string
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

		event, err := parse(raw)
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
			c.lastCursor = event.Cursor
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return decodeErr
}

func (c *Collector) runOnce(ctx context.Context, events chan<- collector.Event) error {
	args := []string{"-f", "-o", "json", "--no-pager"}

	if c.lastCursor != "" {
		args = append(args, "--after-cursor", c.lastCursor)
	}
	cmd := exec.CommandContext(ctx, "journalctl", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

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

	if waitErr != nil {
		return fmt.Errorf("journalctl: %w (%s)", waitErr, strings.TrimSpace(stderr.String()))
	}
	return nil
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
	c.lastCursor = cursor

	delay := time.Second
	shortFailures := 0
	for {
		started := time.Now()
		before := c.lastCursor
		err := c.runOnce(ctx, events)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		advanced := c.lastCursor != before

		if err != nil && !advanced && time.Since(started) < 2*time.Second && c.lastCursor != "" {
			shortFailures++
			if shortFailures >= 3 {
				c.Logger.Warn("journalctl failing immediately with cursor set: clearing cursor, starting from now", "cursor", c.lastCursor, "failures", shortFailures)
				c.lastCursor = ""
				shortFailures = 0
			}
		} else {
			shortFailures = 0
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
