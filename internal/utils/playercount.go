package utils

import (
	"github.com/team-vesperis/vesperis-mp/internal/playerdata"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func GetProxyRealPlayerCount() int {
	return len(p.Players())
}

func GetProxyPlayerCount(player proxy.Player) int {
	var number int

	for _, onlinePlayer := range p.Players() {
		if !playerdata.IsPlayerPrivileged(player) {
			if !playerdata.IsPlayerVanished(onlinePlayer) {
				number++
			}
		} else {
			number++
		}
	}

	return number
}
