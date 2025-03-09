package commands

import (
	"strings"

	"github.com/team-vesperis/vesperis-mp/mp/playerdata"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
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

func getPlayerTarget(playerName string, context *command.Context) proxy.Player {
	player := p.PlayerByName(playerName)

	if player == nil {
		context.SendMessage(&component.Text{
			Content: "Player not found.",
			S: component.Style{
				Color: color.Red,
			},
		})

		return nil
	}

	return player
}

func getPlayerFromSource(source command.Source) proxy.Player {
	player, ok := source.(proxy.Player)

	if !ok {
		return nil
	}

	return player
}

func suggestProxyPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(context *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		remaining := builder.RemainingLowerCase

		players := make([]proxy.Player, 0)
		for _, player := range p.Players() {
			if sourcePlayer, ok := context.Source.(proxy.Player); ok {
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

func suggestAllServers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		return builder.Build()
	})
}

func suggestAllProxies() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		return builder.Build()
	})
}

func suggestAllPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		return builder.Build()
	})
}

func suggestTest() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(ctx *command.Context, builder *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		builder.Suggest("test")
		return builder.Build()
	})
}
