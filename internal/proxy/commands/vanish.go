package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) vanishCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Executes(cm.executeVanishToggle()).
		Requires(cm.requirePrivileged()).
		Then(brigodier.Literal("get").
			Executes(command.Command(func(c *command.Context) error {
				p := cm.getGatePlayerFromSource(c.Source)
				if p == nil {
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

				c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorCyan), "Vanish is: ", v))
				return nil
			})).
			Then(brigodier.Argument("target", brigodier.SingleWord).
				Suggests(cm.suggestAllMultiPlayers(false, false)).
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

					c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorCyan, util.ColorLightGreen), mp.GetUsername()+" is ", v))
					return nil
				})))).
		Then(brigodier.Literal("set").
			Executes(cm.executeVanishToggle()).
			Then(brigodier.Literal("on").
				Executes(cm.executeVanish(true))).
			Then(brigodier.Literal("off").
				Executes(cm.executeVanish(false))))
}

func (cm *CommandManager) executeVanishToggle() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not toggle vanish.", err))
			return err
		}

		v := !mp.IsVanished()

		err = mp.SetVanished(v)
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not toggle vanish.", err))
			return err
		}

		cm.sendVanishMessage(v, c)
		return nil
	})
}

func (cm *CommandManager) executeVanish(v bool) brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not set vanish.", err))
			return err
		}

		if v {
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

		err = mp.SetVanished(v)
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not set vanish.", err))
			return err
		}

		cm.sendVanishMessage(v, c)
		return nil
	})
}

func (cm *CommandManager) sendVanishMessage(v bool, c *command.Context) {
	var m string
	if v {
		m = "on"
	} else {
		m = "off"
	}

	c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorCyan), "Set vanish ", m))
}
