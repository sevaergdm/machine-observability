package collector

import (
	"context"
	"time"
)

const maxFailures = 5

type Collector interface {
	Name() string
	Run(ctx context.Context, out chan<- Event) error
}

type Event interface {
	Source() string
	Timestamp() time.Time
}

type Sampler interface {
	Name() string
	Sample(ctx context.Context) ([]Event, error)
}
