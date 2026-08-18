package wg

import (
	"context"
	"keeper/pkg/eventbus"
	"log/slog"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
)

type EventType int

const (
	ReHandshake EventType = iota
	Handshake
	Timeout
)

type Event struct {
	Type EventType
	Peer Peer
}

type Watcher struct {
	eventBus         *eventbus.EventBus[Event]
	interval         time.Duration
	peerTimeout      time.Duration
	peerNameResolver PeerNameResolver
	device           string
	client           *wgctrl.Client
}

type WatcherOption func(*Watcher)

type state struct {
	lastHandshakes map[string]time.Time
	log            *slog.Logger
}

func NewWatcher(device string, client *wgctrl.Client, opts ...WatcherOption) *Watcher {
	watcher := &Watcher{
		eventBus:         eventbus.New[Event](),
		interval:         5 * time.Second,
		peerTimeout:      3 * time.Minute,
		peerNameResolver: func(_ string) string { return "" },
		device:           device,
		client:           client,
	}

	for _, opt := range opts {
		opt(watcher)
	}

	return watcher
}

func WithInterval(interval time.Duration) WatcherOption {
	return func(w *Watcher) {
		w.interval = interval
	}
}

func WithPeerTimeout(timeout time.Duration) WatcherOption {
	return func(w *Watcher) {
		w.peerTimeout = timeout
	}
}

func WithPeerNameResolver(resolver PeerNameResolver) WatcherOption {
	return func(w *Watcher) {
		w.peerNameResolver = resolver
	}
}

func (w *Watcher) EventBroadcaster() eventbus.Broadcaster[Event] {
	return w.eventBus
}

func (w *Watcher) Run(ctx context.Context) error {
	log := slog.With("component", "WgWatcher", "device", w.device)
	log.Info("polling WireGuard device for events")

	s := state{
		lastHandshakes: make(map[string]time.Time),
		log:            log,
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.tick(s)
		}
	}
}

func (w *Watcher) tick(s state) {
	log, lastHandshakes := s.log, s.lastHandshakes

	dev, err := w.client.Device(w.device)
	if err != nil {
		log.Error("fetching error", "error", err)
		return
	}

	for _, peer := range dev.Peers {
		pubKey := peer.PublicKey.String()
		lastSeen, exists := lastHandshakes[pubKey]

		publishEvent := func(typ EventType) {
			w.eventBus.Publish(Event{
				Type: typ,
				Peer: newPeer(peer, w.peerNameResolver),
			})
		}

		if !peer.LastHandshakeTime.IsZero() && peer.LastHandshakeTime.After(lastSeen) {
			if exists {
				publishEvent(Handshake)
			} else {
				publishEvent(ReHandshake)
			}
			lastHandshakes[pubKey] = peer.LastHandshakeTime
		}

		if exists && time.Since(peer.LastHandshakeTime) > w.peerTimeout {
			delete(lastHandshakes, pubKey)
			publishEvent(Timeout)
		}
	}
}
