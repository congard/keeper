package wg

import (
	"context"
	"fmt"
	"keeper/kit/eventbus"
	"log/slog"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
)

type Peer struct {
	Name     string
	Endpoint string
	PubKey   string
	LastSeen time.Time
}

type PeerNameResolver func(pubkey string) (name string)

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
}

type WatcherOption func(*Watcher)

type state struct {
	client         *wgctrl.Client
	lastHandshakes map[string]time.Time
	log            *slog.Logger
}

func NewWatcher(opts ...WatcherOption) *Watcher {
	watcher := &Watcher{
		eventBus:         eventbus.New[Event](),
		interval:         5 * time.Second,
		peerTimeout:      3 * time.Minute,
		peerNameResolver: func(_ string) string { return "" },
		device:           "wg0",
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

func WithDevice(device string) WatcherOption {
	return func(w *Watcher) {
		w.device = device
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to open wgctrl client: %w", err)
	}
	defer client.Close()

	log := slog.With("component", "WgWatcher", "device", w.device)
	log.Info("polling WireGuard device for events")

	s := state{
		client:         client,
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

	dev, err := s.client.Device(w.device)
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
				Peer: Peer{
					Name:     w.peerNameResolver(pubKey),
					Endpoint: peer.Endpoint.String(),
					PubKey:   pubKey,
					LastSeen: lastSeen,
				},
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
