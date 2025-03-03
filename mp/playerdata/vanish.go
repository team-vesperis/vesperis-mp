package playerdata

import "go.minekube.com/gate/pkg/edition/java/proxy"

func IsPlayerVanished(player proxy.Player) bool {
	playerId := player.ID().String()

	vanished, ok := getPlayerDataField(playerId, "vanished").(bool)
	if !ok {
		vanished = false
		SetPlayerVanished(player, false)
	}

	return vanished
}

func SetPlayerVanished(player proxy.Player, vanished bool) {
	playerId := player.ID().String()

	setPlayerDataField(playerId, "vanished", vanished)
}
