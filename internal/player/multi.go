package player

import "go.minekube.com/common/minecraft/component"

type MultiPlayer struct {
	// The proxy id on which the underlying player is located
    p string

    // The id of the underlying player
    id string
    
    // The username of the underlying player
    name string
}

// New returns a new MultiPlayer
func New() *MultiPlayer {
	return nil
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