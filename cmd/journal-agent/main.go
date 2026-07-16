package main

import (
	"context"
	"fmt"
	"machine-observability/internal/journal"
)

func main() {
	events := make(chan journal.JournalEntry)

	collector := journal.Collector{}

	go func() {
		collector.Run(context.Background(), events)
	}()

	for event := range events {
		fmt.Printf("%+v\n", event)
	}
}
