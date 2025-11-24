package commands

import (
	"strings"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) unBanCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Requires(cm.requireAdminOrModerator()).
		Executes(cm.executeIncorrectUnBanUsage()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Suggests(cm.suggestAllBannedMultiPlayers()).
			Executes(cm.executeUnBan()))
}

func (cm *CommandManager) executeIncorrectUnBanUsage() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		c.SendMessage(util.TextWarn("Incorrect usage: /unban <target>"))
		return nil
	})
}

func (cm *CommandManager) executeUnBan() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}

			c.SendMessage(util.TextInternalError("Could not unban.", err))
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

		c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorCyan), "Unbanned: ", t.GetUsername()))
		return nil
	})
}

func (cm *CommandManager) suggestAllBannedMultiPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase

		if len(r) < 1 {
			b.Suggest("type a username or UUID...")
			return b.Build()
		}

		for _, mp := range cm.mm.GetAllMultiPlayers(false) {
			if !mp.GetBanInfo().IsBanned() {
				continue
			}

			name := mp.GetUsername()
			if strings.HasPrefix(strings.ToLower(name), r) {
				b.Suggest(name)
			}

			id := mp.GetId().String()
			if len(r) > 2 && strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}
