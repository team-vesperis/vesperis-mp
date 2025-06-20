package datasync

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/mp"
)

func IsProxyAvailable(proxyName string) bool {
	proxies, err := GetAllProxies()
	if err == nil {
		return slices.Contains(proxies, proxyName)
	}

	return false
}

func GetProxyWithLowestPlayerCount(countThisProxy bool) (string, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	proxies, err := GetAllProxies()
	if err != nil {
		return "", err
	}

	var minProxy string
	minCount := -1

	for _, proxyName := range proxies {
		if !countThisProxy && proxyName == proxy_name {
			continue
		}

		totalCount := 0
		servers, err := GetServersForProxy(proxyName)
		if err != nil {
			return "", err
		}

		pipe := client.Pipeline()
		for _, serverName := range servers {
			key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
			pipe.HLen(ctx, key)
		}

		cmders, err := pipe.Exec(ctx)
		if err != nil {
			return "", err
		}

		for _, cmder := range cmders {
			totalCount += int(cmder.(*redis.IntCmd).Val())
		}

		if minCount == -1 || totalCount < minCount {
			minCount = totalCount
			minProxy = proxyName
		}
	}

	if minProxy == "" {
		return "", errors.New("no proxies found")
	}

	return minProxy, nil
}

var cache = ttlcache.New[string, []string]()

func GetAllProxies() ([]string, error) {
	key := "proxy_list"
	if cache.Has(key) {
		return cache.Get(key).Value(), nil
	}

	client := database.GetRedisClient()
	ctx := context.Background()

	proxies, err := client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get proxies: %w", err)
	}

	cache.Set(key, proxies, 10*time.Second)
	return proxies, nil
}

func GetAllPlayerUUIDs() ([]string, error) {
	key := "player_uuids_list"
	if cache.Has(key) {
		return cache.Get(key).Value(), nil
	}

	client := database.GetRedisClient()
	ctx := context.Background()

	proxies, err := GetAllProxies()
	if err != nil {
		return nil, err
	}

	var allPlayerUUIDs []string
	for _, proxyName := range proxies {
		servers, err := GetServersForProxy(proxyName)
		if err != nil {
			return nil, err
		}

		pipe := client.Pipeline()
		for _, serverName := range servers {
			key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
			pipe.HKeys(ctx, key)
		}

		cmders, err := pipe.Exec(ctx)
		if err != nil {
			return nil, err
		}

		for _, cmder := range cmders {
			allPlayerUUIDs = append(allPlayerUUIDs, cmder.(*redis.StringSliceCmd).Val()...)
		}
	}

	cache.Set(key, allPlayerUUIDs, 10*time.Second)
	return allPlayerUUIDs, nil
}

func GetAllPlayerNames() ([]string, error) {
	key := "player_names_list"
	if cache.Has(key) {
		list := cache.Get(key).Value()
		return list, nil
	}

	client := database.GetRedisClient()
	ctx := context.Background()

	proxies, err := GetAllProxies()
	if err != nil {
		return nil, err
	}

	var allPlayerNames []string
	for _, proxyName := range proxies {
		servers, err := GetServersForProxy(proxyName)
		if err != nil {
			return nil, err
		}

		for _, serverName := range servers {
			key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
			players, err := client.HGetAll(ctx, key).Result()
			if err != nil {
				return nil, err
			}
			for _, playerName := range players {
				allPlayerNames = append(allPlayerNames, playerName)
			}
		}
	}

	cache.Set(key, allPlayerNames, 1*time.Second)
	return allPlayerNames, nil
}

func GetServersForProxy(proxyName string) ([]string, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	key := fmt.Sprintf("proxy:%s:server_list", proxyName)
	servers, err := client.SMembers(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to get servers for proxy: ", err)
		return nil, err
	}

	return servers, nil
}

func GetPlayersOnServer(proxyName, serverName string) (map[string]string, error) {
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
			if err == redis.Nil {
				continue
			}
			if err != nil {
				return "", "", "", err
			}
			if playerName != "" {
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
			players, err := client.HGetAll(ctx, key).Result()
			if err != nil {
				return "", "", "", err
			}
			for uuid, name := range players {
				if name == playerName {
					return proxyName, serverName, uuid, nil
				}
			}
		}
	}

	return "", "", "", mp.ErrPlayerNotFound
}

func GetTotalPlayerCount() (int, error) {
	client := database.GetRedisClient()
	ctx := context.Background()

	cmd := client.Get(ctx, "player_count")
	if cmd.Err() == redis.Nil {
		client.IncrBy(ctx, "player_count", 0)
		return 0, nil
	}
	count, err := cmd.Int()
	return count, err
}
