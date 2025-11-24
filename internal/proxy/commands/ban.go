package commands

import (
	"errors"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (cm *CommandManager) banCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Requires(cm.requireAdminOrModerator()).
		Executes(cm.executeIncorrectBanUsage()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Executes(cm.executeBan()).
			Suggests(cm.suggestAllMultiPlayers(false, true)).
			Then(brigodier.Argument("reason", brigodier.StringPhrase).
				Executes(cm.executeBan())))
}

func (cm *CommandManager) executeBan() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		r := c.String("reason")
		if r == "" {
			r = "no reason given"
		}

		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}

			c.SendMessage(util.TextInternalError("Could not ban.", err))
			return err
		}

		if !t.IsOnline() {
			err = t.GetBanInfo().Ban(r)
			if err != nil {
				c.SendMessage(util.TextInternalError("Could not ban.", err))
				return err
			}

			c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorCyan), "Banned: ", t.GetUsername(), " Reason: ", r))
			return nil
		}

		mp := t.GetProxy()
		if mp == nil {
			c.SendMessage(util.TextInternalError("Could not ban.", multi.ErrProxyNilWhileOnline))
			return multi.ErrProxyNilWhileOnline
		}

		tr := cm.tm.BuildTask(tasks.NewBanTask(t.GetId(), mp.GetId(), r, true, time.Time{}))
		if !tr.IsSuccessful() {
			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not ban.", err))
			return err
		}

		p, ok := c.Source.(proxy.Player)
		if ok {
			util.PlayThunderSound(p)
		}

		c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorCyan), "Banned: ", t.GetUsername(), " Reason: ", r))
		return nil
	})
}

func (cm *CommandManager) executeIncorrectBanUsage() brigodier.Command {
	return command.Command(func(context *command.Context) error {
		context.SendMessage(util.TextWarn("Incorrect usage: /ban <target> <reason>"))
		return nil
	})
}
