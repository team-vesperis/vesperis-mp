package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (cm *CommandManager) vanishCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Requires(cm.requirePrivileged()).
		Then(brigodier.Literal("get").
			Executes(command.Command(func(c *command.Context) error {
				p, ok := c.Source.(proxy.Player)
				if !ok {
					c.SendMessage(ComponentOnlyPlayersSubCommand)
					return ErrOnlyPlayersSubCommand
				}

				mp, err := cm.mm.GetMultiPlayer(p.ID())
				if err != nil {
					c.SendMessage(util.TextInternalError("Could not get vanish.", err))
					return err
				}

				var text component.Text
				if mp.IsVanished() {
					text = component.Text{
						Content: "active",
						S:       util.StyleColorLightGreen,
					}
				} else {
					text = component.Text{
						Content: "not active",
						S:       util.StyleColorOrange,
					}
				}

				c.SendMessage(&component.Text{
					Content: "Vanish is ",
					S:       util.StyleColorCyan,
					Extra: []component.Component{
						&text,
						&component.Text{
							Content: ".",
							S:       util.StyleColorCyan,
						},
					},
				})

				return nil
			})).
			Then(brigodier.Argument("target", brigodier.SingleWord).
				Suggests(cm.SuggestAllMultiPlayers(false, false)).
				Executes(command.Command(func(c *command.Context) error {
					t := c.String("target")
					mp, err := cm.getMultiPlayerFromTarget(t)
					if err != nil {
						if err == ErrTargetNotFound {
							c.SendMessage(ComponentTargetNotFound)
							return nil
						}
						return err
					}

					var text component.Text
					if mp.IsVanished() {
						text = component.Text{
							Content: "vanished",
							S:       util.StyleColorLightGreen,
						}
					} else {
						text = component.Text{
							Content: "not vanished",
							S:       util.StyleColorOrange,
						}
					}

					c.SendMessage(&component.Text{
						Content: mp.GetUsername() + " is ",
						S:       util.StyleColorCyan,
						Extra: []component.Component{
							&text,
							&component.Text{
								Content: ".",
								S:       util.StyleColorCyan,
							},
						},
					})

					return nil
				})))).
		Then(brigodier.Literal("set").
			Then(brigodier.Literal("on").
				Executes(command.Command(func(c *command.Context) error {
					p, ok := c.Source.(proxy.Player)
					if !ok {
						c.SendMessage(ComponentOnlyPlayersSubCommand)
						return ErrOnlyPlayersSubCommand
					}

					mp, err := cm.mm.GetMultiPlayer(p.ID())
					if err != nil {
						c.SendMessage(util.TextInternalError("Could not set vanish.", err))
						return err
					}

					if mp.IsVanished() {
						c.SendMessage(&component.Text{
							Content: "Vanish is already active.",
							S:       util.StyleColorOrange,
						})
						return nil
					}

					err = mp.SetVanished(true)
					if err != nil {
						c.SendMessage(util.TextInternalError("Could not set vanish.", err))
						return err
					}

					c.SendMessage(&component.Text{
						Content: "Vanish is now active.",
						S:       util.StyleColorLightGreen,
					})

					return nil
				}))).
			Then(brigodier.Literal("off").
				Executes(command.Command(func(c *command.Context) error {
					p, ok := c.Source.(proxy.Player)
					if !ok {
						c.SendMessage(ComponentOnlyPlayersSubCommand)
						return ErrOnlyPlayersSubCommand
					}

					mp, err := cm.mm.GetMultiPlayer(p.ID())
					if err != nil {
						c.SendMessage(util.TextInternalError("Could not set vanish.", err))
						return err
					}

					if !mp.IsVanished() {
						c.SendMessage(&component.Text{
							Content: "Vanish is not active.",
							S:       util.StyleColorOrange,
						})
						return nil
					}

					err = mp.SetVanished(false)
					if err != nil {
						c.SendMessage(util.TextInternalError("Could not set vanish.", err))
						return err
					}

					c.SendMessage(&component.Text{
						Content: "Vanish is now off.",
						S:       util.StyleColorLightGreen,
					})

					return nil
				}))))
}
