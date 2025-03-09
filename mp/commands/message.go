package commands

import (
	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"github.com/team-vesperis/vesperis-mp/mp/mp/task"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func registerMessageCommand() {
	p.Command().Register(messageCommand("message"))
	p.Command().Register(messageCommand("msg"))
}

func messageCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).Then(brigodier.Argument("player", brigodier.SingleWord).Then(brigodier.Argument("message", brigodier.StringPhrase).Executes(sendMessage())))
}

func sendMessage() brigodier.Command {
	return command.Command(func(ctx *command.Context) error {
		player, ok := ctx.Source.(proxy.Player)
		if ok {
			targetName := ctx.String("player")
			target := getPlayerTargetFromThisProxy(targetName, ctx)

			// player is on this server and can be send a normal message
			if target != nil {
				target.SendMessage(&component.Text{
					Content: "[<- " + player.Username() + "]",
					S: component.Style{
						Color: color.Aqua,
					},
					Extra: []component.Component{
						&component.Text{
							Content: ": " + ctx.String("message"),
							S: component.Style{
								Color: color.White,
							},
						},
					},
				})
			} else {
				// player could be on another proxy
				proxyName, _, _, err := datasync.FindPlayerWithUsername(targetName)
				if err.Error() == task.Player_Not_Found {
					player.SendMessage(&component.Text{
						Content: "Player not found.",
						S: component.Style{
							Color: color.Red,
						},
					})
					return nil
				}

				if err != nil {
					player.SendMessage(&component.Text{
						Content: "Error searching player: " + err.Error(),
						S: component.Style{
							Color: color.Red,
						},
					})

					return nil
				}

				messageTask := &task.MessageTask{
					Message:          ctx.String("message"),
					OriginPlayerName: player.Username(),
					TargetPlayerName: targetName,
				}

				err = messageTask.CreateTask(proxyName)
				if err != nil {
					player.SendMessage(&component.Text{
						Content: "Error creating message task: " + err.Error(),
						S: component.Style{
							Color: color.Red,
						},
					})

					return nil
				}

			}

			player.SendMessage(&component.Text{
				Content: "[-> " + targetName + "]",
				S: component.Style{
					Color: color.Aqua,
				},
				Extra: []component.Component{
					&component.Text{
						Content: ": " + ctx.String("message"),
						S: component.Style{
							Color: color.White,
						},
					},
				},
			})
		}

		return nil
	})
}
