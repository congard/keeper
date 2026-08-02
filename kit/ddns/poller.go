package ddns

import (
	"context"
	"log/slog"
	"time"

	"keeper/kit/eventbus"
	"keeper/kit/kitlog"
	"keeper/kit/poller"
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
		kitlog.Error(err, kitlog.WithDescription("ddns update failed"))
		return UpdateResult{}, err
	}

	slog.Info("ddns update succeeded",
		slog.Any("status", result.Status),
		kitlog.OptionalString("ipv4", result.IPv4),
		kitlog.OptionalString("ipv6", result.IPv6),
	)

	return result, nil
}

var _ poller.Worker[UpdateResult] = updaterWorker{}
