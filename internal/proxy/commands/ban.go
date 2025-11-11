package commands

import (
	"errors"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) banCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Requires(cm.requireAdminOrModerator()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Executes(cm.executeBan()).
			Suggests(cm.SuggestAllMultiPlayers(false, true)).
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
			return err
		}

		if !t.IsOnline() {
			c.SendMessage(TextTargetIsOffline)
			return nil
		}

		mp := t.GetProxy()
		if mp == nil {
			c.SendMessage(util.TextInternalError("Could not ban.", multi.ErrProxyNilWhileOnline))
			return multi.ErrProxyNilWhileOnline
		}

		tr := cm.tm.BuildTask(tasks.NewBanTask(t.GetId(), mp.GetId(), r, true, time.Now()))
		if !tr.IsSuccessful() {
			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not ban.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors("Banned: ", t.GetUsername(), " Reason: ", r))
		return nil
	})
}
