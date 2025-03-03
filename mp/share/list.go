package share

import (
	"context"
	"strings"

	"github.com/team-vesperis/vesperis-mp/mp/database"
)

func GetAllPlayerNames() []string {
	var list []string
	client := database.GetRedisClient()
	ctx := context.Background()

	playerNames, err := client.SMembers(ctx, "player_list").Result()
	if err != nil {
		logger.Error("Error getting player names: ", err)
		return list
	}

	return playerNames
}

func GetAllServerNames() []string {
	var list []string
	client := database.GetRedisClient()
	ctx := context.Background()

	serverNames, err := client.SMembers(ctx, "server_list").Result()
	if err != nil {
		logger.Error("Error getting server names: ", err)
		return list
	}

	return serverNames
}

func GetAllServerNamesFromProxy(proxyName string) []string {
	var list []string
	client := database.GetRedisClient()
	ctx := context.Background()

	serverNames, err := client.SMembers(ctx, "server_list").Result()
	if err != nil {
		logger.Error("Error getting server names: ", err)
		return list
	}

	for _, serverName := range serverNames {
		if strings.HasPrefix(serverName, proxyName+"_") {
			list = append(list, serverName)
		}
	}

	return list
}

func GetAllProxyNames() []string {
	var list []string
	client := database.GetRedisClient()
	ctx := context.Background()

	serverNames, err := client.SMembers(ctx, "proxy_list").Result()
	if err != nil {
		logger.Error("Error getting proxy names: ", err)
		return list
	}

	return serverNames
}

func AddPlayer(playerName string) {
	client := database.GetRedisClient()
	ctx := context.Background()

	err := client.SAdd(ctx, "player_list", playerName).Err()
	if err != nil {
		logger.Error("Error adding player: ", err)
	}
}

func RemovePlayer(playerName string) {
	client := database.GetRedisClient()
	ctx := context.Background()

	err := client.SRem(ctx, "player_list", playerName).Err()
	if err != nil {
		logger.Error("Error removing player: ", err)
	}
}
