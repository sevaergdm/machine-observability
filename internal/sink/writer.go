package sink

import (
	"context"
	"log/slog"
	"machine-observability/internal/collector"
	"time"
)

const (
	DefaultMaxRows = 10000
	DefaultMaxAge  = 60 * time.Second
)

type FlushFunc func(rows []collector.Event) error

type Writer struct {
	events     <-chan collector.Event
	rowBuffer  []collector.Event
	maxRows    int
	maxAge     time.Duration
	flush      FlushFunc
	logger     *slog.Logger
	firstRowAt time.Time
}

type Tuning struct {
	MaxRows int
	MaxAge  time.Duration
}

func NewWriter(events <-chan collector.Event, flush FlushFunc, maxRows int, maxAge time.Duration, logger *slog.Logger) *Writer {
	return &Writer{
		events:    events,
		rowBuffer: make([]collector.Event, 0, maxRows),
		maxRows:   maxRows,
		maxAge:    maxAge,
		flush:     flush,
		logger:    logger,
	}
}

func (w *Writer) doFlush() {
	if len(w.rowBuffer) == 0 {
		return
	}

	if err := w.flush(w.rowBuffer); err != nil {
		// Flush failure will not clear the rowBuffer or reset firstRowsAt
		// TODO: Handle a maximum amount of retries to prevent an infinte retry loop
		w.logger.Error("flush failed", "error", err)
		return
	}
	w.rowBuffer = make([]collector.Event, 0, w.maxRows)
	w.firstRowAt = time.Time{}
}

func (w *Writer) Run(ctx context.Context) error {
	if w.logger == nil {
		w.logger = slog.New(slog.DiscardHandler)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case event, ok := <-w.events:
			if !ok {
				w.logger.Debug("final flush", "rows", len(w.rowBuffer))
				w.doFlush()
				return nil
			}

			if len(w.rowBuffer) == 0 {
				w.firstRowAt = time.Now()
			}
			w.rowBuffer = append(w.rowBuffer, event)
			if len(w.rowBuffer) >= w.maxRows {
				w.logger.Debug("flushing on row count", "rows", len(w.rowBuffer))
				w.doFlush()
			}
		case <-ticker.C:
			if len(w.rowBuffer) > 0 && time.Since(w.firstRowAt) >= w.maxAge {
				w.logger.Debug("flushing on max age", "age", time.Since(w.firstRowAt).Seconds())
				w.doFlush()
			}
		case <-ctx.Done():
			w.doFlush()
			return ctx.Err()
		}
	}
}
