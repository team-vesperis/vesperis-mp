package commands

import (
	"github.com/team-vesperis/vesperis-mp/mp/mp/transfer"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func registerTransferCommand() {
	p.Command().Register(transferCommand())
}

func transferCommand() brigodier.LiteralNodeBuilder {
	return brigodier.Literal("transfer").
		Then(brigodier.Argument("proxy", brigodier.String).
			Executes(command.Command(func(context *command.Context) error {
				player, ok := context.Source.(proxy.Player)
				if ok {
					proxy := context.String("proxy")
					err := transfer.TransferPlayerToProxy(player, proxy)
					if err != nil {
						player.SendMessage(&component.Text{
							Content: "Error transferring: " + err.Error(),
							S:       component.Style{Color: color.Red},
						})
						return nil
					}
				}

				return nil
			})).
			Suggests(suggestAllProxies()).
			Then(brigodier.Argument("server", brigodier.String).
				Executes(command.Command(func(ctx *command.Context) error {
					player, ok := ctx.Source.(proxy.Player)
					if ok {
						proxy := ctx.String("proxy")
						server := ctx.String("server")
						err := transfer.TransferPlayerToServerOnOtherProxy(player, proxy, server)
						if err != nil {
							player.SendMessage(&component.Text{
								Content: "Error transferring: " + err.Error(),
								S:       component.Style{Color: color.Red},
							})
							return nil
						}
					}

					return nil
				})).Suggests(suggestAllServersFromProxy()).
				Then(brigodier.Argument("player", brigodier.String).
					Executes(command.Command(func(ctx *command.Context) error {
						player := p.PlayerByName(ctx.String("player"))
						if player == nil {

							return nil
						}

						proxy := ctx.String("proxy")
						server := ctx.String("server")
						err := transfer.TransferPlayerToServerOnOtherProxy(player, proxy, server)
						if err != nil {
							player.SendMessage(&component.Text{
								Content: "Error transferring " + player.Username() + ": " + err.Error(),
								S: component.Style{
									Color: color.Red,
								},
							})
						}

						return nil
					})).
					Suggests(suggestAllPlayers())))).
		Requires(requireAdminOrModerator())
}
