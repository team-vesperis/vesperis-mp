package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) kickCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Suggests(cm.SuggestAllMultiPlayers(true)).
			Then(brigodier.Argument("reason", brigodier.StringPhrase).
				Executes(command.Command(func(c *command.Context) error {
					mp, err := cm.getMultiPlayerFromTarget(c.String("target"))
					if err != nil {
						if err == ErrTargetNotFound {
							c.SendMessage(ComponentTargetNotFound)
							return nil
						}
						return err
					}

					if !mp.IsOnline() {
						c.SendMessage(ComponentTargetIsOffline)
						return nil
					}

					mproxy := mp.GetProxy()
					if mproxy == nil {
						return nil
					}

					tr := cm.tm.BuildTask(tasks.NewKickTask(mp.GetId(), mp.GetProxy().GetId(), c.String("reason")))
					if !tr.IsSuccessful() {
						err := errors.New(tr.GetReason())
						c.SendMessage(util.TextInternalError("Could not kick.", err))
						return err
					}

					c.SendMessage(&component.Text{
						Content: "Kicked: ",
						S:       util.StyleColorLightGreen,
						Extra: []component.Component{
							&component.Text{
								Content: mp.GetUsername(),
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
