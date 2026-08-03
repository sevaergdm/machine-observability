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
	events     chan collector.Event
	rowBuffer  []collector.Event
	maxRows    int
	maxAge     time.Duration
	flush      FlushFunc
	logger     *slog.Logger
	firstRowAt time.Time
}

func NewWriter(flush FlushFunc, maxRows int, maxAge time.Duration, logger *slog.Logger) *Writer {
	return &Writer{
		events:    make(chan collector.Event),
		rowBuffer: make([]collector.Event, 0, maxRows),
		maxRows:   maxRows,
		maxAge:    maxAge,
		flush:     flush,
		logger:    logger,
	}
}

func (w *Writer) doFlush() error {
	if len(w.rowBuffer) == 0 {
		return nil
	}

	if err := w.flush(w.rowBuffer); err != nil {
		return err
	}
	w.rowBuffer = make([]collector.Event, 0, w.maxRows)
	w.firstRowAt = time.Time{}
	return nil
}

func (w *Writer) Run(ctx context.Context) error {
	if w.logger == nil {
		w.logger = slog.New(slog.DiscardHandler)
	}

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case event, ok := <-w.events:
			if !ok {
				w.logger.Debug("nothing to flush", "rows", len(w.rowBuffer))
				err := w.doFlush()
				if err != nil {
					w.logger.Error("flush failed", "error", err)
				}
			}

			if len(w.rowBuffer) == 0 {
				w.firstRowAt = time.Now()
			}
			w.rowBuffer = append(w.rowBuffer, event)
			if len(w.rowBuffer) >= w.maxRows {
				w.logger.Debug("flushing on row count", "rows", len(w.rowBuffer))
				err := w.doFlush()
				if err != nil {
					w.logger.Error("flush failed", "error", err)
				}
			}
		case <-ticker.C:
			if len(w.rowBuffer) > 0 && time.Since(w.firstRowAt) >= w.maxAge {
				w.logger.Debug("flushing on max age", "age", time.Since(w.firstRowAt).Seconds())
				err := w.doFlush()
				if err != nil {
					w.logger.Error("flush failed", "error", err)
				}
			}
		case <-ctx.Done():
			err := w.doFlush()
			if err != nil {
				w.logger.Error("flush failed", "error", err)
			}
		}
	}
}
