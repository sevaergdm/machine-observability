package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"machine-observability/internal/collector"
	"machine-observability/internal/config"
	"machine-observability/internal/journal"
	"machine-observability/internal/sink"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type buildDeps struct {
	logger   *slog.Logger
	stateDir string
}

type registration struct {
	kind  config.Kind
	build func(d buildDeps) collector.Collector
}

var registry = map[string]registration{
	"journal": {
		kind: config.Streaming,
		build: func(d buildDeps) collector.Collector {
			return &journal.Collector{
				Logger:     d.logger,
				CursorPath: filepath.Join(d.stateDir, "journal.cursor"),
			}
		},
	},
}

func main() {
	var level slog.LevelVar
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: &level}))

	configPath := flag.String("config", "", "path to the config.toml")
	flag.Parse()
	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "error: no config file specified\n")
		flag.Usage()
		os.Exit(2)
	}

	knownNames := make(map[string]config.Kind)
	for name, reg := range registry {
		knownNames[name] = reg.kind
	}

	cfg, err := config.Load(*configPath, knownNames)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	level.Set(cfg.LogLevel)

	logger.Info("agent starting", "config", *configPath, "data_dir", cfg.DataDir, "state_dir", cfg.StateDir, "log_level", cfg.LogLevel)

	err = os.MkdirAll(cfg.DataDir, 0750)
	if err != nil {
		logger.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	err = os.MkdirAll(cfg.StateDir, 0750)
	if err != nil {
		logger.Error("failed to create state directory", "error", err)
		os.Exit(1)
	}

	err = sink.CleanTmp(cfg.DataDir, logger)
	if err != nil {
		logger.Error("failed to clean data directory", "error", err)
	}

	var active []collector.Collector
	var names []string
	for name, collectorConfig := range cfg.Collectors {
		if !collectorConfig.Enabled {
			continue
		}
		active = append(active, registry[name].build(buildDeps{
			logger:   logger.With("collector", name),
			stateDir: cfg.StateDir,
		}))
		names = append(names, name)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// events is deliberately unbuffered: collectors block while the sink is mid-flush
	// Acceptable because flushes are fast local-disk writes and journald buffers upstream
	// Revisit if flush latency ever grows
	events := make(chan collector.Event)

	g, ctx := errgroup.WithContext(ctx)
	for _, c := range active {
		g.Go(func() error { return c.Run(ctx, events) })
	}
	logger.Info("collectors enabled", "collectors", names, "count", len(names))

	go func() {
		if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("collector failed with error", "error", err)
		}
		close(events)
	}()

	parquetFlush := sink.NewParquetFlush[journal.Entry](cfg.DataDir, "journal")
	flushFn := func(rows []collector.Event) error {
		if err := parquetFlush(rows); err != nil {
			return err
		}
		// safe: parquetFlush validated row types above
		last := rows[len(rows)-1].(journal.Entry)
		return last.WriteCursor(cfg.StateDir)
	}

	w := sink.NewWriter(events, flushFn, sink.DefaultMaxRows, sink.DefaultMaxAge, logger.With("component", "sink"))
	// Using background context here because reusing the existing ctx would mean that the writer would cancel while
	// messages are still being pushed into events during shutdown. By using the separate context, we ensure that
	// the writer stays until the channel closes
	if err := w.Run(context.Background()); err != nil {
		logger.Error("sink exited with error", "error", err)
	}
	logger.Info("shutdown complete")
}
