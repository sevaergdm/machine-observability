package main

import (
	"context"
	"fmt"
	"log"
	"machine-observability/internal/collector"
	"machine-observability/internal/config"
	"machine-observability/internal/journal"
)

var CONFIG_PATH = "../../config.toml"

var knownNames = map[string]bool{
	"journal": true,
}

func main() {
	events := make(chan collector.Event)

	collector := journal.Collector{}

	config, err := config.Load(CONFIG_PATH, knownNames)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		collector.Run(context.Background(), events)
	}()

	for event := range events {
		fmt.Printf("%+v\n", event)
	}
}
