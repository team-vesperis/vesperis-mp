package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) unBanCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Requires(cm.requireAdminOrModerator()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Suggests(cm.SuggestAllBannedMultiPlayers()).
			Executes(cm.executeUnBan()))
}

func (cm *CommandManager) executeUnBan() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}
			return err
		}

		if !t.GetBanInfo().IsBanned() {
			c.SendMessage(util.TextWarn("Target is not banned."))
			return nil
		}

		err = t.GetBanInfo().UnBan()
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not unban.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors("Unbanned: ", t.GetUsername()))
		return nil
	})
}
