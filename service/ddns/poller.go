package ddns

import (
	"context"
	"log/slog"
	"time"

	"keeper/pkg/eventbus"
	"keeper/pkg/logger"
	"keeper/pkg/poller"
)

type Subscriber eventbus.Subscriber[UpdateResult]

type Poller struct {
	*poller.Poller[UpdateResult]
}

func New(interval time.Duration) *Poller {
	return &Poller{
		Poller: poller.New[UpdateResult](interval),
	}
}

func (p *Poller) AddUpdater(updater Updater) {
	p.AddWorker(updaterWorker{updater})
}

type updaterWorker struct {
	updater Updater
}

func (w updaterWorker) Do(_ context.Context) (UpdateResult, error) {
	result, err := w.updater.Update()
	if err != nil {
		logger.LogIfError(err, logger.WithDescription("ddns update failed"))
		return UpdateResult{}, err
	}

	slog.Info("ddns update succeeded",
		slog.Any("status", result.Status),
		logger.OptionalString("ipv4", result.IPv4),
		logger.OptionalString("ipv6", result.IPv6),
	)

	return result, nil
}

var _ poller.Worker[UpdateResult] = updaterWorker{}
