package telemetry

import (
	"context"
	"time"

	"keeper/pkg/poller"
)

type Poller struct {
	*poller.Poller[HostMetrics]
}

func New(interval time.Duration) *Poller {
	return &Poller{
		Poller: poller.New[HostMetrics](interval),
	}
}

func (p *Poller) AddCollector(collector Collector) {
	p.AddWorker(collectorWorker{collector})
}

type collectorWorker struct {
	collector Collector
}

func (w collectorWorker) Do(ctx context.Context) (HostMetrics, error) {
	return w.collector.Collect(ctx)
}

var _ poller.Worker[HostMetrics] = collectorWorker{}
