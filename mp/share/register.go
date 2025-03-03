package share

import (
	"context"

	"github.com/team-vesperis/vesperis-mp/mp/database"
	"go.minekube.com/common/minecraft/component"
)

func RegisterServer(server_name string) {
	client := database.GetRedisClient()
	ctx := context.Background()

	register_name := proxy_name + "_" + server_name
	exists, err := client.SIsMember(ctx, "server_list", register_name).Result()
	if err != nil {
		logger.Error("Failed to check if server exists: ", err)
	}

	if exists {
		logger.Error("Server already registered: ", register_name)
		p.Shutdown(&component.Text{Content: "Server already registered: " + register_name})
	}

	err = client.SAdd(ctx, "server_list", register_name).Err()
	if err != nil {
		logger.Error("Failed to register server", register_name, ": ", err)
	}

	logger.Info("Registered server: ", register_name)
}

func UnregisterServer(server_name string) {
	client := database.GetRedisClient()
	ctx := context.Background()

	register_name := proxy_name + "_" + server_name

	err := client.SRem(ctx, "server_list", register_name).Err()
	if err != nil {
		logger.Error("Failed to unregister server ", register_name, ": ", err)
	}

	logger.Info("Unregistered server: ", register_name)
}

func registerServers() {
	for _, server := range p.Servers() {
		server_name := server.ServerInfo().Name()
		RegisterServer(server_name)
	}
}

func unregisterServers() {
	client := database.GetRedisClient()
	ctx := context.Background()

	for _, server := range p.Servers() {
		server_name := server.ServerInfo().Name()
		register_name := proxy_name + "_" + server_name

		err := client.SRem(ctx, "server_list", register_name).Err()
		if err != nil {
			logger.Error("Failed to unregister server ", register_name, ": ", err)
		}

		logger.Info("Unregistered server: ", register_name)
	}
}

func registerProxy() {
	client := database.GetRedisClient()
	ctx := context.Background()

	exists, err := client.SIsMember(ctx, "proxy_list", proxy_name).Result()
	if err != nil {
		logger.Error("Failed to check if proxy exists: ", err)
	}

	if exists {
		logger.Error("Proxy already registered: ", proxy_name)
		p.Shutdown(&component.Text{Content: "Proxy already registered: " + proxy_name})
	}

	err = client.SAdd(ctx, "proxy_list", proxy_name).Err()
	if err != nil {
		logger.Error("Failed to register proxy: ", err)
	}

	logger.Info("Registered proxy: ", proxy_name)
}

func unregisterProxy() {
	client := database.GetRedisClient()
	ctx := context.Background()

	err := client.SRem(ctx, "proxy_list", proxy_name).Err()
	if err != nil {
		logger.Error("Failed to unregister proxy: ", err)
	}

	logger.Info("Unregistered proxy: ", proxy_name)
}
