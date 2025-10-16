package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/util/uuid"

	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (cm *CommandManager) messageCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Suggests(cm.SuggestAllMultiPlayers(true)).
			Then(brigodier.Argument("message", brigodier.StringPhrase).
				Executes(command.Command(func(c *command.Context) error {
					t, err := cm.getMultiPlayerFromTarget(c.String("target"))
					if err != nil {
						if err == ErrTargetNotFound {
							c.SendMessage(ComponentTargetNotFound)
							return nil
						}

						c.SendMessage(util.TextInternalError("Could not send message.", err))
						return err
					}

					if !t.IsOnline() {
						c.SendMessage(ComponentTargetIsOffline)
						return nil
					}

					mproxy := t.GetProxy()
					if mproxy == nil {
						c.SendMessage(util.TextInternalError("Could not send message.", multi.ErrProxyNilWhileOnline))
						return multi.ErrProxyNilWhileOnline
					}

					if mproxy.GetId() == uuid.Nil {
						c.SendMessage(util.TextInternalError("Could not send message.", multi.ErrProxyIdNilWhileOnline))
						return multi.ErrProxyIdNilWhileOnline
					}

					var originName string
					p, ok := c.Source.(proxy.Player)
					if ok {
						mp, err := cm.mpm.GetMultiPlayer(p.ID())
						if err != nil {
							c.SendMessage(util.TextInternalError("Could not send message.", err))
							return err
						}

						originName = mp.GetNickname()
					} else {
						originName = "Vesperis-Proxy-" + cm.mpm.GetOwnerMultiProxy().GetId().String()
					}

					tr := cm.tm.BuildTask(tasks.NewMessageTask(originName, t.GetId(), t.GetProxy().GetId(), c.String("message")))
					if !tr.IsSuccessful() {
						err := errors.New(tr.GetReason())
						c.SendMessage(util.TextInternalError("Could not send message.", err))
						return err
					}

					c.SendMessage(&component.Text{
						Content: "[-> " + t.GetUsername() + "]",
						S:       util.StyleColorCyan,
						Extra: []component.Component{
							&component.Text{
								Content: ": " + c.String("message"),
								S: component.Style{
									Color: color.White,
								},
							},
						},
					})

					return nil
				}))))
}
