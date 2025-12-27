package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (cm *CommandManager) refreshCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Requires(cm.requireAdmin()).
		Executes(cm.executeRefresh(false)).
		Then(brigodier.Argument("proxyId", brigodier.SingleWord).
			Suggests(cm.suggestAllMultiProxies(false)).
			Executes(cm.executeRefresh(true)))
}

func (cm *CommandManager) executeRefresh(withTarget bool) brigodier.Command {
	return command.Command(func(c *command.Context) error {
		var proxyId uuid.UUID
		if withTarget {
			proxyString := c.String("proxyId")
			var err error
			proxyId, err = uuid.Parse(proxyString)
			if err != nil {
				c.SendMessage(util.TextWarn("Invalid Proxy UUID"))
				return nil
			}

			_, err = cm.mm.GetMultiProxy(proxyId)
			if err != nil {
				if err == database.ErrDataNotFound {
					c.SendMessage(util.TextWarn("Proxy not found."))
					return nil
				}

				c.SendMessage(util.TextInternalError("Could not refresh.", err))
				return err
			}

		} else {
			proxyId = cm.mm.GetOwnerMultiProxy().GetId()
		}

		tr := cm.tm.BuildTask(tasks.NewRefreshTask(proxyId))
		if !tr.IsSuccessful() {
			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not refresh.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorOrange, util.ColorLightBlue), "Successfully refreshed proxy.", "Duration: ", tr.GetInfo()))
		return nil
	})
}
