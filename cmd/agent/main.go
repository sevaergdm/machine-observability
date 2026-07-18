package main

import (
	"context"
	"fmt"
	"machine-observability/internal/journal"
	"machine-observability/internal/collector"
)

func main() {
	events := make(chan collector.Event)

	collector := journal.Collector{}

	go func() {
		collector.Run(context.Background(), events)
	}()

	for event := range events {
		fmt.Printf("%+v\n", event)
	}
}
