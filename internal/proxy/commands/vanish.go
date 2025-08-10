package commands

import (
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (cm *CommandManager) vanishCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
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
						Content: "Could not turn on vanish: error getting multiplayer.",
						S: component.Style{
							Color: color.Red,
						},
					})
					return err
				}

				if mp.IsVanished() {
					c.SendMessage(&component.Text{
						Content: "Vanish is already active.",
						S:       component.Style{},
					})
					return nil
				}

				err = mp.SetVanished(true, true)
				if err != nil {

				}

				return nil
			}))).
			Then(brigodier.Literal("off")))
}
