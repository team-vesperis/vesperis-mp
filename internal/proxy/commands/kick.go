package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) kickCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Executes(cm.executeIncorrectKick()).
		Requires(cm.requireAdminOrModerator()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Executes(cm.executeKick()).
			Suggests(cm.SuggestAllMultiPlayers(true, true)).
			Then(brigodier.Argument("reason", brigodier.StringPhrase).
				Executes(cm.executeKick())))
}

func (cm *CommandManager) executeIncorrectKick() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		c.SendMessage(util.TextWarn("Incorrect usage: /kick <target> <reason>"))
		return nil
	})
}

func (cm *CommandManager) executeKick() brigodier.Command {
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
			c.SendMessage(util.TextInternalError("Could not kick.", multi.ErrProxyNilWhileOnline))
			return multi.ErrProxyNilWhileOnline
		}

		tr := cm.tm.BuildTask(tasks.NewKickTask(t.GetId(), mp.GetId(), r))
		if !tr.IsSuccessful() {
			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not kick.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors("Kicked: ", t.GetUsername(), " Reason: ", r))
		return nil

	})
}
