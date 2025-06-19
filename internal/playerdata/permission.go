package playerdata

import (
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var validRoles = map[string]struct{}{
	"admin":     {},
	"builder":   {},
	"default":   {},
	"moderator": {},
}

var validRanks = map[string]struct{}{
	"champion": {},
	"default":  {},
	"elite":    {},
	"legend":   {},
}

func GetPlayerRole(player proxy.Player) string {
	playerId := player.ID().String()

	roleInterface := getPlayerDataField(playerId, "role")
	role := "default"
	if roleInterface == nil {
		SetPlayerRole(player, "default")
	} else {
		role = roleInterface.(string)
	}

	_, valid := validRoles[role]
	if !valid {
		role = "default"
		SetPlayerRole(player, "default")
	}

	return role
}

func SetPlayerRole(player proxy.Player, role string) {
	playerId := player.ID().String()

	_, valid := validRoles[role]
	if !valid {
		logger.Error("Invalid role: ", role)
		return
	}

	setPlayerDataField(playerId, "role", role)
	logger.Info("Changed permission role for " + player.Username() + " - " + playerId + " to " + role)
}

func IsPlayerPrivileged(player proxy.Player) bool {
	role := GetPlayerRole(player)
	return role == "admin" || role == "builder" || role == "moderator"
}

func GetPlayerRank(player proxy.Player) string {
	playerId := player.ID().String()

	rankInterface := getPlayerDataField(playerId, "rank")
	rank := "default"
	if rankInterface == nil {
		SetPlayerRank(player, "default")
	} else {
		rank = rankInterface.(string)
	}

	_, valid := validRanks[rank]
	if !valid {
		rank = "default"
		SetPlayerRank(player, "default")
	}

	return rank
}

func SetPlayerRank(player proxy.Player, rank string) {
	playerId := player.ID().String()

	_, valid := validRanks[rank]
	if !valid {
		logger.Error("Invalid rank: ", rank)
		return
	}

	setPlayerDataField(playerId, "rank", rank)
	logger.Info("Changed permission rank for " + player.Username() + " - " + playerId + " to " + rank)
}
