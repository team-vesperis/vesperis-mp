package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/multiproxy/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (cm *CommandManager) vanishCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("get").Executes(command.Command(func(c *command.Context) error {
			p, ok := c.Source.(proxy.Player)
			if !ok {
				c.SendMessage(ComponentOnlyPlayersSubCommand)
				return ErrOnlyPlayersSubCommand
			}

			mp, err := cm.mpm.GetMultiPlayer(p.ID())
			if err != nil {
				c.SendMessage(&component.Text{
					Content: "Could not get vanish.",
					S: component.Style{
						Color:      util.ColorRed,
						HoverEvent: component.ShowText(&component.Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
					},
				})
				return err
			}

			if mp.IsVanished() {
				c.SendMessage(&component.Text{
					Content: "Vanish is active.",
					S:       util.StyleColorLightGreen,
				})
				return nil
			}

			c.SendMessage(&component.Text{
				Content: "Vanish is not active.",
				S:       util.StyleColorLightGreen,
			})
			return nil
		})).Then(brigodier.Argument("target", brigodier.SingleWord))).
		Then(brigodier.Literal("set").
			Then(brigodier.Literal("on").Executes(command.Command(func(c *command.Context) error {
				p, ok := c.Source.(proxy.Player)
				if !ok {
					c.SendMessage(ComponentOnlyPlayersSubCommand)
					return ErrOnlyPlayersSubCommand
				}

				mp, err := cm.mpm.GetMultiPlayer(p.ID())
				if err != nil {
					c.SendMessage(&component.Text{
						Content: "Could not turn on vanish.",
						S: component.Style{
							Color:      util.ColorRed,
							HoverEvent: component.ShowText(&component.Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
						},
					})
					return err
				}

				if mp.IsVanished() {
					c.SendMessage(&component.Text{
						Content: "Vanish is already active.",
						S:       util.StyleColorOrange,
					})
					return nil
				}

				err = mp.SetVanished(true, true)
				if err != nil {
					c.SendMessage(&component.Text{
						Content: "Could not turn on vanish.",
						S: component.Style{
							Color:      util.ColorRed,
							HoverEvent: component.ShowText(&component.Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
						},
					})
					return err
				}

				return nil
			}))).
			Then(brigodier.Literal("off").Executes(command.Command(func(c *command.Context) error {
				p, ok := c.Source.(proxy.Player)
				if !ok {
					c.SendMessage(ComponentOnlyPlayersSubCommand)
					return ErrOnlyPlayersSubCommand
				}

				mp, err := cm.mpm.GetMultiPlayer(p.ID())
				if err != nil {
					c.SendMessage(&component.Text{
						Content: "Could not turn off vanish: error getting multiplayer.",
						S: component.Style{
							Color:      util.ColorRed,
							HoverEvent: component.ShowText(&component.Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
						},
					})
					return err
				}

				if !mp.IsVanished() {
					c.SendMessage(&component.Text{
						Content: "Vanish is not active.",
						S:       util.StyleColorOrange,
					})
					return nil
				}

				err = mp.SetVanished(false, true)
				if err != nil {
					c.SendMessage(&component.Text{
						Content: "Could not turn off vanish: error getting multiplayer.",
						S: component.Style{
							Color:      util.ColorRed,
							HoverEvent: component.ShowText(&component.Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
						},
					})
					return err
				}

				return nil
			}))))
}
