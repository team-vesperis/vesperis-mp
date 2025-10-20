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
			Suggests(cm.SuggestAllMultiPlayers(true, true)).
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

					mp := t.GetProxy()
					if mp == nil {
						c.SendMessage(util.TextInternalError("Could not send message.", multi.ErrProxyNilWhileOnline))
						return multi.ErrProxyNilWhileOnline
					}

					if mp.GetId() == uuid.Nil {
						c.SendMessage(util.TextInternalError("Could not send message.", multi.ErrProxyIdNilWhileOnline))
						return multi.ErrProxyIdNilWhileOnline
					}

					var originName string
					p, ok := c.Source.(proxy.Player)
					if ok {
						mp, err := cm.mm.GetMultiPlayer(p.ID())
						if err != nil {
							c.SendMessage(util.TextInternalError("Could not send message.", err))
							return err
						}

						if t == mp {
							c.SendMessage(&component.Text{
								Content: "You can't message yourself. You can add friends using /friends add <player_name>",
								S:       util.StyleColorOrange,
							})

							return nil
						}

						// hide vanished from non-privileged players
						if !mp.GetPermissionInfo().IsPrivileged() {
							if t.IsVanished() {
								c.SendMessage(ComponentTargetIsOffline)
								return nil
							}
						} else {
							if mp.IsVanished() && !t.GetPermissionInfo().IsPrivileged() {
								c.SendMessage(&component.Text{
									Content: "Warning: You are in vanish but your sending a message to a non-privileged player. This player can not respond. Turn off vanish to message correctly with non-privileged players.",
									S:       util.StyleColorOrange,
								})
							}
						}

						originName = mp.GetNickname()
					} else {
						originName = "Vesperis-Proxy-" + cm.mm.GetOwnerMultiProxy().GetId().String()
					}

					tr := cm.tm.BuildTask(tasks.NewMessageTask(originName, t.GetId(), t.GetProxy().GetId(), c.String("message")))
					if !tr.IsSuccessful() {
						err := errors.New(tr.GetInfo())
						c.SendMessage(util.TextInternalError("Could not send message.", err))
						return err
					}

					c.SendMessage(&component.Text{
						Content: "[-> " + t.GetUsername() + "]",
						S: component.Style{
							Color: util.ColorCyan,
							HoverEvent: component.ShowText(&component.Text{
								Content: "Click to send another message",
								S:       util.StyleColorGray,
							}),
							ClickEvent: component.SuggestCommand("/message " + t.GetUsername() + " "),
						},
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
