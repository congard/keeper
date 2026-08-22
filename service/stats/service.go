package stats

import (
	"context"
	"keeper/pkg/eventbus"
	"keeper/pkg/svcutil"
	"log/slog"
	"time"
)

type Service struct {
	context      context.Context
	pollInterval time.Duration
	eventBus     *eventbus.EventBus[HostMetrics]
	logger       *slog.Logger
}

func NewService(ctx context.Context, pollInterval time.Duration) *Service {
	logger := slog.With("component", "StatsService")

	service := &Service{
		pollInterval: pollInterval,
		context:      ctx,
		eventBus:     eventbus.New[HostMetrics](),
		logger:       logger,
	}

	svcutil.Go(logger, service.runPoller)

	return service
}

func (s *Service) EventBroadcaster() eventbus.Broadcaster[HostMetrics] {
	return s.eventBus
}

func (s *Service) runPoller() error {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.context.Done():
			return s.context.Err()
		case <-ticker.C:
			s.tickPoller()
		}
	}
}

func (s *Service) tickPoller() {
	timeoutContext, cancel := context.WithTimeout(s.context, 2*time.Second)
	defer cancel()

	stats, err := Collect(timeoutContext)
	if err != nil {
		s.logger.Error("failed to collect host metrics", "error", err)
		return
	}

	s.eventBus.Publish(stats)
}
