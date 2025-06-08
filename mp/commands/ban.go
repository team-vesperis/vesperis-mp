package commands

import (
	"github.com/team-vesperis/vesperis-mp/mp/ban"
	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"github.com/team-vesperis/vesperis-mp/mp/mp/task"
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
		targetName := context.String("player")
		target := p.PlayerByName(targetName)
		if target != nil {
			playerRole := playerdata.GetPlayerRole(target)
			if playerRole == "admin" || playerRole == "moderator" {
				context.SendMessage(&component.Text{
					Content: "You are not allowed to ban admins or moderators. If an admin or moderator is not following the rules, you can join the discord for help.",
					S: component.Style{
						Color: color.Red,
					},
				})

				return nil
			}

			reason := context.String("reason")
			ban.BanPlayer(target, reason)

			return nil
		}

		// player could be on another proxy
		proxyName, _, _, err := datasync.FindPlayerWithUsername(targetName)
		if err == task.ErrPlayerNotFound {
			context.SendMessage(&component.Text{
				Content: "Player not found.",
				S: component.Style{
					Color: color.Red,
				},
			})
			return nil
		}

		if err != task.ErrSuccessful {
			context.SendMessage(&component.Text{
				Content: "Error searching player: " + err.Error(),
				S: component.Style{
					Color: color.Red,
				},
			})

			return nil
		}

		banTask := &task.BanTask{
			TargetPlayerName: targetName,
			Reason:           context.String("reason"),
		}

		err = banTask.CreateTask(proxyName)
		if err != nil {
			context.SendMessage(&component.Text{
				Content: "Error creating ban task: " + err.Error(),
				S: component.Style{
					Color: color.Red,
				},
			})
		}

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
