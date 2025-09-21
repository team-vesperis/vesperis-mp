package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) messageCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Then(brigodier.Argument("message", brigodier.StringPhrase).
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

					cm.tm.SendTask(mp.GetProxy(), &tasks.MessageTask{})

					// handle normally
					if mproxy.GetId() == cm.mpm.GetOwnerProxyId() {

					}

					return nil
				}))))
}
