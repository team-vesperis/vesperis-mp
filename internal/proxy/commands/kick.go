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
							c.SendMessage(ComponentTargetNotFound)
							return nil
						}
						return err
					}

					if !t.IsOnline() {
						c.SendMessage(ComponentTargetIsOffline)
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

					c.SendMessage(&component.Text{
						Content: "Kicked: ",
						S:       util.StyleColorLightGreen,
						Extra: []component.Component{
							&component.Text{
								Content: t.GetUsername(),
								S:       util.StyleColorCyan,
							},
							&component.Text{
								Content: ". Reason: ",
								S:       util.StyleColorLightGreen,
							},
							&component.Text{
								Content: c.String("reason"),
								S:       util.StyleColorCyan,
							},
						},
					})

					return nil
				}))))
}
