package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (cm *CommandManager) transferCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Suggests(cm.SuggestAllMultiPlayers(true, false)).
			Then(brigodier.Argument("proxyId", brigodier.SingleWord).
				Suggests(cm.SuggestAllMultiProxies(false)).
				Executes(cm.transfer()).
				Then(brigodier.Argument("backendId", brigodier.SingleWord).
					Suggests(cm.SuggestAllMultiBackendsUnderProxy(true)).
					Executes(cm.transfer()))))
}

func (cm *CommandManager) transfer() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		proxyString := c.String("proxyId")
		proxyId, err := uuid.Parse(proxyString)
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not transfer.", err))
		}

		backendString := c.String("backendId")
		backendId := uuid.Nil

		if backendString != "" {
			backendId, err = uuid.Parse(backendString)
			if err != nil {
				c.SendMessage(util.TextInternalError("Could not transfer.", err))
			}
		}

		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(ComponentTargetNotFound)
				return nil
			}

			c.SendMessage(util.TextInternalError("Could not transfer.", err))
			return err
		}

		if !t.IsOnline() {
			c.SendMessage(ComponentTargetIsOffline)
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
			c.SendMessage(&component.Text{
				Content: "Target is already on that backend.",
				S:       util.StyleColorOrange,
			})
			return nil
		}

		if backendId == uuid.Nil && mp.GetId() == proxyId {
			c.SendMessage(&component.Text{
				Content: "Target is already on that proxy.",
				S:       util.StyleColorOrange,
			})
			return nil
		}

		tr := cm.tm.BuildTask(tasks.NewTransferTask(t.GetId(), mp.GetId(), proxyId, backendId))

		if !tr.IsSuccessful() {
			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not transfer.", err))
			return err
		}

		return nil
	})
}
