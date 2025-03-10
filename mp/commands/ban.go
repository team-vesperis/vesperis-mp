package commands

import (
	"github.com/team-vesperis/vesperis-mp/mp/ban"
	"github.com/team-vesperis/vesperis-mp/mp/playerdata"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
)

func registerBanCommand() {
	p.Command().Register(banCommand())
	logger.Info("Registered ban command.")
}

func banCommand() brigodier.LiteralNodeBuilder {
	return brigodier.Literal("ban").
		Executes(incorrectBanCommandUsage()).
		Then(brigodier.Argument("player", brigodier.SingleWord).
			Executes(incorrectBanCommandUsage()).
			Suggests(suggestAllPlayers()).
			Then(brigodier.Argument("reason", brigodier.StringPhrase).
				Executes(banPlayer()))).
		Requires(requireAdminOrModerator())
}

func banPlayer() brigodier.Command {
	return command.Command(func(context *command.Context) error {
		playerName := context.String("player")
		player := getPlayerTargetFromThisProxy(playerName, context)
		if player == nil {
			return nil
		}

		if playerdata.GetPlayerRole(player) == "admin" {
			context.SendMessage(&component.Text{
				Content: "You are not allowed to ban admins.",
				S: component.Style{
					Color: color.Red,
				},
			})

			return nil
		}

		reason := context.String("reason")
		ban.BanPlayer(player, reason)

		return nil
	})
}

func incorrectBanCommandUsage() brigodier.Command {
	return command.Command(func(context *command.Context) error {
		context.SendMessage(&component.Text{
			Content: "Incorrect usage: /ban <player> <reason>",
			S: component.Style{
				Color: color.Red,
			},
		})
		return nil
	})
}
