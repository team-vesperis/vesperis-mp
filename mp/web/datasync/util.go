package datasync

import (
	"context"
	"errors"
	"fmt"

	"github.com/team-vesperis/vesperis-mp/mp/database"
)

func GetAllProxies() ([]string, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	proxies, err := client.SMembers(ctx, "proxies").Result()
	if err != nil {
		logger.Error("Failed to get proxies: ", err)
		return nil, err
	}

	return proxies, nil
}

func GetServersForProxy(proxyName string) ([]string, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	key := fmt.Sprintf("proxy:%s:servers", proxyName)
	servers, err := client.SMembers(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get servers for proxy: ", err)
		return nil, err
	}

	return servers, nil
}

func GetPlayersForServer(proxyName, serverName string) (map[string]string, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
	players, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get players for server: ", err)
		return nil, err
	}

	return players, nil
}

func FindPlayerWithUUID(playerUUID string) (string, string, string, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	proxies, err := GetAllProxies()
	if err != nil {
		return "", "", "", err
	}

	for _, proxyName := range proxies {
		servers, err := GetServersForProxy(proxyName)
		if err != nil {
			return "", "", "", err
		}

		for _, serverName := range servers {
			key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
			playerName, err := client.HGet(ctx, key, playerUUID).Result()
			if err == nil && playerName != "" {
				return proxyName, serverName, playerName, nil
			}
		}
	}

	return "", "", "", errors.New("player not found")
}

func FindPlayerWithUsername(playerName string) (string, string, string, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	proxies, err := GetAllProxies()
	if err != nil {
		return "", "", "", err
	}

	for _, proxyName := range proxies {
		servers, err := GetServersForProxy(proxyName)
		if err != nil {
			return "", "", "", err
		}

		for _, serverName := range servers {
			key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
			playerUUID, err := client.HGet(ctx, key, playerName).Result()
			if err == nil && playerUUID != "" {
				return proxyName, serverName, playerUUID, nil
			}
		}
	}

	return "", "", "", errors.New("player not found")
}

func GetTotalPlayerCount() (int, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	proxies, err := GetAllProxies()
	if err != nil {
		return 0, err
	}

	totalCount := 0
	for _, proxyName := range proxies {
		servers, err := GetServersForProxy(proxyName)
		if err != nil {
			return 0, err
		}

		for _, serverName := range servers {
			key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
			playerCount, err := client.HLen(ctx, key).Result()
			if err != nil {
				return 0, err
			}
			totalCount += int(playerCount)
		}
	}

	return totalCount, nil
}
