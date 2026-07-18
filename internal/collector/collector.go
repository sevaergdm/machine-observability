package collector

import (
	"context"
	"time"
)

type Collector interface {
	Name() string
	Run(ctx context.Context, out chan<- Event) error
}

type Event interface {
	Source() string
	Timestamp() time.Time
}
