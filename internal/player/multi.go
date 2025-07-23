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
    return mp.p
}

func (mp *MultiPlayer) Id() string {
    return mp.id
}

func (mp *MultiPlayer) Name() string {
    return mp.name
}