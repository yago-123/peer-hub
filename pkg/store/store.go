package store

import "github.com/yago-123/peer-hub/pkg/peer"

type Store interface {
	Register(peerID string, info peer.Info) error
	Lookup(peerID string) (peer.Info, bool)
}
