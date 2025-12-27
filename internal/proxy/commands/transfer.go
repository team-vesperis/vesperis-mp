package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (cm *CommandManager) transferCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Executes(cm.executeIncorrectUsage("/transfer <target> <proxyId> <backendId>")).
		Requires(cm.requireAdmin()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Executes(cm.executeIncorrectUsage("/transfer <target> <proxyId> <backendId>")).
			Suggests(cm.suggestAllMultiPlayers(true, false)).
			Then(brigodier.Argument("proxyId", brigodier.SingleWord).
				Suggests(cm.suggestAllMultiProxies(false)).
				Executes(cm.executeTransfer(false)).
				Then(brigodier.Argument("backendId", brigodier.SingleWord).
					Suggests(cm.suggestAllMultiBackendsUnderProxy(true)).
					Executes(cm.executeTransfer(true)))))
}

func (cm *CommandManager) executeTransfer(withBackend bool) brigodier.Command {
	return command.Command(func(c *command.Context) error {
		proxyString := c.String("proxyId")
		proxyId, err := uuid.Parse(proxyString)
		if err != nil {
			c.SendMessage(util.TextWarn("Invalid Proxy UUID"))
			return nil
		}

		backendId := uuid.Nil
		if withBackend {
			backendString := c.String("backendId")

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

		if !withBackend && mp.GetId() == proxyId {
			c.SendMessage(util.TextWarn("Target is already on that proxy."))
			return nil
		}

		tr := cm.tm.BuildTask(tasks.NewTransferTask(t.GetId(), mp.GetId(), proxyId, backendId))

		if !tr.IsSuccessful() {
			if tr.GetInfo() == tasks.ErrStringProxyNotFound {
				c.SendMessage(util.TextWarn("Proxy not found."))
				return nil
			}

			if tr.GetInfo() == tasks.ErrStringBackendNotFound {
				c.SendMessage(util.TextWarn("Backend not found."))
				return nil
			}

			if tr.GetInfo() == util.ErrStringBackendNotResponding {
				c.SendMessage(util.TextWarn("Backend was found but is not responding."))
				return nil
			}

			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not transfer.", err))
			return err
		}

		return nil
	})
}
