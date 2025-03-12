package commands

import (
	"errors"
	"strings"

	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"github.com/team-vesperis/vesperis-mp/mp/playerdata"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var p *proxy.Proxy
var logger *zap.SugaredLogger

func InitializeCommands(proxy *proxy.Proxy, log *zap.SugaredLogger) {
	p = proxy
	logger = log

	registerTransferCommand()
	registerBanCommand()
	registerTempBanCommand()
	registerUnBanCommand()
	registerPermissionCommand()
	registerMessageCommand()

	brigodier.ErrDispatcherUnknownArgument = errors.New("Unknown argument.")

	logger.Info("Successfully registered all commands.")
}

func requireAdmin() brigodier.RequireFn {
	return command.Requires(func(context *command.RequiresContext) bool {
		player := getPlayerFromSource(context.Source)

		if player != nil {
			return playerdata.GetPlayerRole(player) == "admin"
		}

		return false
	})
}

func requireAdminOrModerator() brigodier.RequireFn {
	return command.Requires(func(context *command.RequiresContext) bool {
		player := getPlayerFromSource(context.Source)

		if player != nil {
			permission := playerdata.GetPlayerRole(player)
			return permission == "admin" || permission == "moderator"
		}

		return false

	})
}

func requireAdminOrBuilder() brigodier.RequireFn {
	return command.Requires(func(context *command.RequiresContext) bool {
		player := getPlayerFromSource(context.Source)

		if player != nil {
			permission := playerdata.GetPlayerRole(player)
			return permission == "admin" || permission == "builder"
		}

		return false

	})
}

func requireStaff() brigodier.RequireFn {
	return command.Requires(func(context *command.RequiresContext) bool {
		player := getPlayerFromSource(context.Source)

		if player != nil {
			return playerdata.IsPlayerPrivileged(player)
		}

		return false
	})
}

func getPlayerFromSource(source command.Source) proxy.Player {
	player, ok := source.(proxy.Player)

	if !ok {
		return nil
	}

	return player
}

func suggestProxyPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		remaining := builder.RemainingLowerCase

		players := make([]proxy.Player, 0)
		for _, player := range p.Players() {
			if sourcePlayer, ok := ctx.Source.(proxy.Player); ok {
				if playerdata.IsPlayerPrivileged(sourcePlayer) || playerdata.IsPlayerVanished(player) {
					if strings.HasPrefix(strings.ToLower(player.Username()), remaining) {
						players = append(players, player)
					}
				}
			} else {
				if strings.HasPrefix(strings.ToLower(player.Username()), remaining) {
					players = append(players, player)
				}
			}
		}

		if len(players) != 0 {
			for _, player := range players {
				builder.Suggest(player.Username())
			}
		}

		return builder.Build()
	})
}

func suggestProxyServers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		remaining := builder.RemainingLowerCase

		servers := make([]proxy.RegisteredServer, 0)
		for _, server := range p.Servers() {
			if strings.HasPrefix(strings.ToLower(server.ServerInfo().Name()), remaining) {
				servers = append(servers, server)
			}
		}

		if len(servers) != 0 {
			for _, server := range servers {
				builder.Suggest(server.ServerInfo().Name())
			}
		}

		return builder.Build()
	})
}

func suggestAllServersFromProxy() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		remaining := builder.RemainingLowerCase
		server_list, err := datasync.GetServersForProxy(ctx.String("proxy"))
		if err != nil {
			logger.Error("Error getting all server names from one proxy: ", err)
		}

		servers := make([]string, 0)
		for _, server := range server_list {
			if strings.HasPrefix(strings.ToLower(server), remaining) {
				servers = append(servers, server)
			}
		}

		if len(servers) != 0 {
			for _, server := range servers {
				builder.Suggest(server)
			}
		}

		return builder.Build()
	})
}

func suggestAllProxies() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		remaining := builder.RemainingLowerCase
		proxy_list, err := datasync.GetAllProxies()
		if err != nil {
			logger.Error("Error getting all proxy names: ", err)
		}

		proxies := make([]string, 0)
		for _, proxy := range proxy_list {
			if strings.HasPrefix(strings.ToLower(proxy), remaining) {
				proxies = append(proxies, proxy)
			}
		}

		if len(proxies) != 0 {
			for _, proxy := range proxies {
				builder.Suggest(proxy)
			}
		}

		return builder.Build()
	})
}

func suggestAllPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		remaining := builder.RemainingLowerCase
		player_list, err := datasync.GetAllPlayerNames()
		if err != nil {
			logger.Error("Error getting all player names: ", err)
		}

		players := make([]string, 0)
		for _, player := range player_list {
			if strings.HasPrefix(strings.ToLower(player), remaining) {
				players = append(players, player)
			}
		}

		if len(players) != 0 {
			for _, player := range players {
				builder.Suggest(player)
			}
		}

		return builder.Build()
	})
}

func suggestTest() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		builder.Suggest("test")
		return builder.Build()
	})
}
