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

func (cm *CommandManager) kickCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Requires(cm.requireAdminOrModerator()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Suggests(cm.SuggestAllMultiPlayers(true, true)).
			Then(brigodier.Argument("reason", brigodier.StringPhrase).
				Executes(command.Command(func(c *command.Context) error {
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
						c.SendMessage(util.TextInternalError("Could not send message.", multi.ErrProxyNilWhileOnline))
						return multi.ErrProxyNilWhileOnline
					}

					if mp.GetId() == uuid.Nil {
						c.SendMessage(util.TextInternalError("Could not send message.", multi.ErrProxyIdNilWhileOnline))
						return multi.ErrProxyIdNilWhileOnline
					}

					tr := cm.tm.BuildTask(tasks.NewKickTask(t.GetId(), t.GetProxy().GetId(), c.String("reason")))
					if !tr.IsSuccessful() {
						err := errors.New(tr.GetInfo())
						c.SendMessage(util.TextInternalError("Could not kick.", err))
						return err
					}

					c.SendMessage(util.TextSuccess("Kicked: ", t.GetUsername(), " Reason: ", c.String("reason")))
					return nil
				}))))
}
