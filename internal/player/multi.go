package player

import (
    "go.minekube.com/common/minecraft/component"
    "go.minekube.com/gate/pkg/edition/java/proxy"
)

type MultiPlayer struct {
    // The proxy id on which the underlying player is located
    p string

    // The id of the underlying player
    id string
    
    // The username of the underlying player
    name string

    mu sync.RWMutex
}

// New returns a new MultiPlayer
func New(p proxy.Player, id string) *MultiPlayer {
	return &MultiPlayer{
        p: id,
        id: p.ID().String(),
        name: p.Username()
    }
}

func (mp *MultiPlayer) ProxyId() string {
    mp.mu.RLock()
    defer mp.mu.RUnlock()
    return mp.p
}

func (mp *MultiPlayer) Id() string {
    mp.mu.RLock()
    defer mp.mu.RUnlock()
    return mp.id
}

func (mp *MultiPlayer) Name() string {
    mp.mu.RLock()
    defer mp.mu.RUnlock()
    return mp.name
}