package datasync

import (
	"context"
	"fmt"

	"github.com/team-vesperis/vesperis-mp/mp/database"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func registerProxy(proxyName string) error {
	client := database.GetRedisClient()
	ctx := context.Background()

	exists, err := client.SIsMember(ctx, "proxies", proxyName).Result()
	if err != nil {
		logger.Error("Failed to check if proxy is already registered: ", err)
		return err
	}

	if exists {
		logger.Error("Proxy already registered: ", proxyName)
		p.Shutdown(&component.Text{
			Content: "Proxy already registered.",
		})
		return nil
	}

	err = client.SAdd(ctx, "proxies", proxyName).Err()
	if err != nil {
		logger.Error("Failed to register proxy: ", err)
		return err
	}

	logger.Info("Registered proxy: ", proxyName)
	return nil
}

func unregisterProxy(proxyName string) error {
	client := database.GetRedisClient()
	ctx := context.Background()

	// Remove all servers under the proxy
	serverKey := fmt.Sprintf("proxy:%s:servers", proxyName)
	servers, err := client.SMembers(ctx, serverKey).Result()
	if err != nil {
		logger.Error("Failed to get servers for proxy: ", err)
		return err
	}

	for _, server := range servers {
		err := UnregisterServer(proxyName, server)
		if err != nil {
			logger.Error("Failed to unregister server: ", server, " for proxy: ", err)
			return err
		}
	}

	// Remove the proxy itself
	err = client.SRem(ctx, "proxies", proxyName).Err()
	if err != nil {
		logger.Error("Failed to unregister proxy: ", err)
		return err
	}

	logger.Info("Unregistered proxy: ", proxyName)
	return nil
}

func RegisterServer(proxyName, serverName string) error {
	client := database.GetRedisClient()
	ctx := context.Background()

	key := fmt.Sprintf("proxy:%s:servers", proxyName)
	err := client.SAdd(ctx, key, serverName).Err()
	if err != nil {
		logger.Error("Failed to register server: ", err)
		return err
	}

	logger.Info("Registered server: ", serverName, " under proxy: ", proxyName)
	return nil
}

func UnregisterServer(proxyName, serverName string) error {
	client := database.GetRedisClient()
	ctx := context.Background()

	key := fmt.Sprintf("proxy:%s:servers", proxyName)
	err := client.SRem(ctx, key, serverName).Err()
	if err != nil {
		logger.Error("Failed to unregister server: ", err)
		return err
	}

	logger.Info("Unregistered server: ", serverName, " from proxy: ", proxyName)
	return nil
}

func RegisterPlayer(proxyName, serverName string, player proxy.Player) error {
	client := database.GetRedisClient()
	ctx := context.Background()

	key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
	err := client.HSet(ctx, key, player.ID().String(), player.Username()).Err()
	if err != nil {
		logger.Error("Failed to register player: ", err)
		return err
	}

	logger.Info("Registered player: ", player.Username(), " (UUID: ", player.ID().String(), ") under server: ", serverName, " and proxy: ", proxyName)
	return nil
}

func UnregisterPlayer(proxyName, serverName string, player proxy.Player) error {
	client := database.GetRedisClient()
	ctx := context.Background()

	key := fmt.Sprintf("proxy:%s:server:%s:players", proxyName, serverName)
	err := client.HDel(ctx, key, player.ID().String()).Err()
	if err != nil {
		logger.Error("Failed to unregister player: ", err)
		return err
	}

	logger.Info("Unregistered player with UUID: ", player.ID().String(), " from server: ", serverName, " and proxy: ", proxyName)
	return nil
}
