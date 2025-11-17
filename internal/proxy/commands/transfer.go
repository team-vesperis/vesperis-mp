package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (cm *CommandManager) transferCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Executes(cm.executeIncorrectTransfer()).
		Requires(cm.requireAdmin()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Executes(cm.executeIncorrectTransfer()).
			Suggests(cm.SuggestAllMultiPlayers(true, false)).
			Then(brigodier.Argument("proxyId", brigodier.SingleWord).
				Suggests(cm.SuggestAllMultiProxies(false)).
				Executes(cm.executeTransfer()).
				Then(brigodier.Argument("backendId", brigodier.SingleWord).
					Suggests(cm.SuggestAllMultiBackendsUnderProxy(true)).
					Executes(cm.executeTransfer()))))
}

func (cm *CommandManager) executeIncorrectTransfer() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		c.SendMessage(util.TextWarn("Incorrect usage: /transfer <target> <proxyId> <backendId>"))
		return nil
	})
}

func (cm *CommandManager) executeTransfer() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		proxyString := c.String("proxyId")
		proxyId, err := uuid.Parse(proxyString)
		if err != nil {
			c.SendMessage(util.TextWarn("Invalid Proxy UUID"))
			return nil
		}

		backendString := c.String("backendId")
		backendId := uuid.Nil

		if backendString != "" {
			backendId, err = uuid.Parse(backendString)
			if err != nil {
				c.SendMessage(util.TextWarn("Invalid Backend UUID"))
				return nil
			}
		}

		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}

			c.SendMessage(util.TextInternalError("Could not transfer.", err))
			return err
		}

		if !t.IsOnline() {
			c.SendMessage(TextTargetIsOffline)
			return nil
		}

		mp := t.GetProxy()
		if mp == nil {
			c.SendMessage(util.TextInternalError("Could not transfer.", multi.ErrProxyNilWhileOnline))
			return multi.ErrProxyNilWhileOnline
		}

		mb := t.GetBackend()
		if mb == nil {
			c.SendMessage(util.TextInternalError("Could not transfer.", multi.ErrProxyNilWhileOnline))
			return multi.ErrBackendNilWhileOnline
		}

		if mb.GetId() == backendId {
			c.SendMessage(util.TextWarn("Target is already on that backend."))
			return nil
		}

		if backendId == uuid.Nil && mp.GetId() == proxyId {
			c.SendMessage(util.TextWarn("Target is already on that proxy."))
			return nil
		}

		tr := cm.tm.BuildTask(tasks.NewTransferTask(t.GetId(), mp.GetId(), proxyId, backendId))

		if !tr.IsSuccessful() {
			if tr.GetInfo() == database.ErrDataNotFound.Error() {
				c.SendMessage(util.TextWarn("Proxy not found."))
			}
			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not transfer.", err))
			return err
		}

		return nil
	})
}
