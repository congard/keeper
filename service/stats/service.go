package stats

import (
	"context"
	"keeper/pkg/eventbus"
	"keeper/pkg/logger"
	"log/slog"
	"time"
)

type Service struct {
	pollInterval time.Duration
	eventBus     *eventbus.EventBus[HostMetrics]
	logger       *slog.Logger
}

func NewService(pollInterval time.Duration) *Service {
	return &Service{
		pollInterval: pollInterval,
		eventBus:     eventbus.New[HostMetrics](),
		logger:       logger.ForComponent[Service](),
	}
}

func (s *Service) EventBroadcaster() eventbus.Broadcaster[HostMetrics] {
	return s.eventBus
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Service) tick(ctx context.Context) {
	timeoutContext, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	stats, err := Collect(timeoutContext)
	if err != nil {
		s.logger.Error("failed to collect host metrics", "error", err)
		return
	}

	s.eventBus.Publish(stats)
}
