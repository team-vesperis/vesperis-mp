package player

import "go.minekube.com/common/minecraft/component"

type MultiPlayer interface {
	// The proxy id on which the underlying player is located
	Proxy() string

	// Send a message to the player
	SendMessage(component.Text) error
}

// New returns a new MultiPlayer that can bred
func New() *MultiPlayer {
	return nil
}
