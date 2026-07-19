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
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type registration struct {
	kind  config.Kind
	build func() collector.Collector
}

var registry = map[string]registration{
	"journal": {
		kind:  config.Streaming,
		build: func() collector.Collector { return &journal.Collector{} },
	},
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

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

	var active []collector.Collector
	for name, collectorConfig := range cfg.Collectors {
		if !collectorConfig.Enabled {
			continue
		}
		active = append(active, registry[name].build())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	events := make(chan collector.Event)

	g, ctx := errgroup.WithContext(ctx)
	for _, c := range active {
		g.Go(func() error { return c.Run(ctx, events) })
	}

	go func() {
		if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("collector failed with error", "error", err)
		}
		close(events)
	}()

	for event := range events {
		fmt.Printf("%+v\n", event)
	}
}
