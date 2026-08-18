package wg

import (
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Peer struct {
	Name          string
	Endpoint      string
	PubKey        string
	LastSeen      time.Time
	ReceiveBytes  int64
	TransmitBytes int64
}

type PeerNameResolver func(pubkey string) (name string)

func newPeer(p wgtypes.Peer, nameResolver PeerNameResolver) Peer {
	pubKey := p.PublicKey.String()
	return Peer{
		Name:          nameResolver(pubKey),
		Endpoint:      p.Endpoint.String(),
		PubKey:        pubKey,
		LastSeen:      p.LastHandshakeTime,
		ReceiveBytes:  p.ReceiveBytes,
		TransmitBytes: p.TransmitBytes,
	}
}
