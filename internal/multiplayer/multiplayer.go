package multiplayer

import (
	"go.minekube.com/common/minecraft/component"
)

type MultiPlayer interface {
	// The proxy id on which the underlying player is located
	Proxy() string

	// Send a message to the player
	SendMessage(component.Text) error
}
