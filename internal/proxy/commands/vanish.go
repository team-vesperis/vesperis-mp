package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
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

				var v string
				if mp.IsVanished() {
					v = "active"
				} else {
					v = "not active"
				}

				c.SendMessage(util.TextAlternatingColors("Vanish is: ", v))
				return nil
			})).
			Then(brigodier.Argument("target", brigodier.SingleWord).
				Suggests(cm.SuggestAllMultiPlayers(false, false)).
				Executes(command.Command(func(c *command.Context) error {
					t := c.String("target")
					mp, err := cm.getMultiPlayerFromTarget(t)
					if err != nil {
						if err == ErrTargetNotFound {
							c.SendMessage(TextTargetNotFound)
							return nil
						}
						return err
					}

					var v string
					if mp.IsVanished() {
						v = "vanished"
					} else {
						v = "not vanished"
					}

					c.SendMessage(util.TextAlternatingColors(mp.GetUsername()+" is ", v))
					return nil
				})))).
		Then(brigodier.Literal("set").
			Then(brigodier.Literal("on").
				Executes(cm.executeVanish(true))).
			Then(brigodier.Literal("off").
				Executes(cm.executeVanish(false))))
}

func (cm *CommandManager) executeVanish(vanish bool) brigodier.Command {
	return command.Command(func(c *command.Context) error {
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

		if vanish {
			if mp.IsVanished() {
				c.SendMessage(util.TextWarn("Vanish is already active"))
				return nil
			}

		} else {
			if !mp.IsVanished() {
				c.SendMessage(util.TextWarn("Vanish is not active"))
				return nil
			}
		}

		err = mp.SetVanished(vanish)
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not set vanish.", err))
			return err
		}

		if vanish {
			c.SendMessage(util.TextAlternatingColors("Vanish is now active"))
		} else {
			c.SendMessage(util.TextAlternatingColors("Vanish is now not active"))
		}

		return nil
	})
}
