package sink

import (
	"context"
	"fmt"
	"log/slog"
	"machine-observability/internal/collector"
	"sync"
)

type route struct {
	ch chan collector.Event
	w  *Writer
}

type Manager struct {
	events  <-chan collector.Event
	routes  map[string]route
	wg      sync.WaitGroup
	logger  *slog.Logger
	started bool
}

func NewManager(events <-chan collector.Event, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	return &Manager{
		events: events,
		routes: make(map[string]route),
		logger: logger,
	}
}

func (m *Manager) Register(source string, flush FlushFunc, t Tuning) error {
	if t.MaxRows <= 0 || t.MaxAge <= 0 {
		return fmt.Errorf("invalid tuning for %q", source)
	}

	if _, exists := m.routes[source]; exists {
		return fmt.Errorf("source %q is already registered", source)
	}

	if m.started {
		return fmt.Errorf("register %q after Run started", source)
	}

	ch := make(chan collector.Event)
	w := NewWriter(ch, flush, t.MaxRows, t.MaxAge, m.logger.With("source", source))
	m.routes[source] = route{ch: ch, w: w}

	return nil
}

func (m *Manager) Run() {

	m.started = true

	for source, r := range m.routes {
		m.wg.Go(func() {
			if err := r.w.Run(context.Background()); err != nil {
				m.logger.Error("writer exited with error", "error", err, "source", source)
			}
		})
	}

	for event := range m.events {
		r, ok := m.routes[event.Source()]
		if !ok {
			m.logger.Error("event with unregistered source dropped", "source", event.Source())
			continue
		}
		r.ch <- event
	}

	for _, r := range m.routes {
		close(r.ch)
	}
	m.wg.Wait()
}
