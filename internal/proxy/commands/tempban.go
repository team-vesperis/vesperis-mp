package commands

import (
	"errors"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) tempBanCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Executes(incorrectTempBanCommandUsage()).
		Then(brigodier.Argument("target", brigodier.SingleWord).
			Executes(incorrectTempBanCommandUsage()).
			Suggests(cm.SuggestAllMultiPlayers(false, true)).
			Then(brigodier.Argument("time_amount", brigodier.Int).
				Executes(incorrectTempBanCommandUsage()).
				Then(brigodier.Literal("seconds").
					Executes(cm.executeTempBan(time.Second)).
					Then(brigodier.Argument("reason", brigodier.StringPhrase).
						Executes(cm.executeTempBan(time.Second)))).
				Then(brigodier.Literal("minutes").
					Executes(cm.executeTempBan(time.Minute)).
					Then(brigodier.Argument("reason", brigodier.StringPhrase).
						Executes(cm.executeTempBan(time.Minute)))).
				Then(brigodier.Literal("hours").
					Executes(cm.executeTempBan(time.Hour)).
					Then(brigodier.Argument("reason", brigodier.StringPhrase).
						Executes(cm.executeTempBan(time.Hour)))).
				Then(brigodier.Literal("days").
					Executes(cm.executeTempBan(time.Hour * 24)).
					Then(brigodier.Argument("reason", brigodier.StringPhrase).
						Executes(cm.executeTempBan(time.Hour * 24)))))).
		Requires(cm.requireAdminOrModerator())
}

func (cm *CommandManager) executeTempBan(time_type time.Duration) brigodier.Command {
	return command.Command(func(c *command.Context) error {
		r := c.String("reason")
		if r == "" {
			r = "no reason given"
		}

		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}
			return err
		}

		d := time_type * time.Duration(c.Int("time_amount"))
		e := util.FormatDuration(d)
		expiration := time.Now().Add(d)

		if !t.IsOnline() {
			err = t.GetBanInfo().TempBan(r, expiration)
			if err != nil {
				c.SendMessage(util.TextInternalError("Could not tempban.", err))
				return err
			}

			c.SendMessage(util.TextAlternatingColors("Temp banned: ", t.GetUsername(), "\nReason: ", r, "\nExpiration: ", e))
			return nil
		}

		mp := t.GetProxy()
		if mp == nil {
			c.SendMessage(util.TextInternalError("Could not tempban.", multi.ErrProxyNilWhileOnline))
			return multi.ErrProxyNilWhileOnline
		}

		tr := cm.tm.BuildTask(tasks.NewBanTask(t.GetId(), mp.GetId(), r, false, expiration))
		if !tr.IsSuccessful() {
			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not tempban.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors("Temp banned: ", t.GetUsername(), "\nReason: ", r, "\nExpiration: ", e))
		return nil
	})
}

func incorrectTempBanCommandUsage() brigodier.Command {
	return command.Command(func(context *command.Context) error {
		context.SendMessage(&component.Text{
			Content: "Incorrect usage: /tempban <player> <time_amount> <time_type> <reason>",
			S: component.Style{
				Color: color.Red,
			},
		})
		return nil
	})
}
