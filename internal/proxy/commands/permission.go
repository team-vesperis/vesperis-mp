package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) permissionCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("set").
			Then(brigodier.Literal("role").
				Then(brigodier.Argument("target", brigodier.SingleWord).
					Then(brigodier.Literal(multi.RoleAdmin.String()).
						Executes(cm.executeSetRole(multi.RoleAdmin))).
					Then(brigodier.Literal(multi.RoleBuilder.String()).
						Executes(cm.executeSetRole(multi.RoleBuilder))).
					Then(brigodier.Literal(multi.RoleDefault.String()).
						Executes(cm.executeSetRole(multi.RoleDefault))).
					Then(brigodier.Literal(multi.RoleModerator.String()).
						Executes(cm.executeSetRole(multi.RoleModerator))).
					Executes(cm.executeIncorrectPermissionCommandSetUsage()).
					Suggests(cm.SuggestAllMultiPlayers(false, false))).
				Executes(cm.executeIncorrectPermissionCommandSetUsage())).
			Executes(cm.executeIncorrectPermissionCommandSetUsage()).
			Then(brigodier.Literal("rank").
				Then(brigodier.Argument("target", brigodier.SingleWord).
					Then(brigodier.Literal(multi.RankChampion.String()).
						Executes(cm.executeSetRank(multi.RankChampion))).
					Then(brigodier.Literal(multi.RankDefault.String()).
						Executes(cm.executeSetRank(multi.RankDefault))).
					Then(brigodier.Literal(multi.RankElite.String()).
						Executes(cm.executeSetRank(multi.RankElite))).
					Then(brigodier.Literal(multi.RankLegend.String()).
						Executes(cm.executeSetRank(multi.RankLegend))).
					Executes(cm.executeIncorrectPermissionCommandSetUsage()).
					Suggests(cm.SuggestAllMultiPlayers(false, false))).
				Executes(cm.executeIncorrectPermissionCommandSetUsage())).
			Requires(cm.requireAdmin())).
		Then(brigodier.Literal("get").
			Then(brigodier.Literal("rank").
				Then(brigodier.Argument("target", brigodier.SingleWord).
					Executes(cm.executeGetRank()).
					Suggests(cm.SuggestAllMultiPlayers(false, false)))).
			Then(brigodier.Literal("role").
				Then(brigodier.Argument("target", brigodier.SingleWord).
					Executes(cm.executeGetRole()).
					Suggests(cm.SuggestAllMultiPlayers(false, false)))).
			Executes(cm.executeIncorrectPermissionCommandGetUsage())).
		Executes(cm.executeIncorrectPermissionCommandUsage()).
		Requires(cm.requireAdminOrModerator())
}

func (cm *CommandManager) executeSetRole(r multi.Role) brigodier.Command {
	return command.Command(func(c *command.Context) error {
		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not set role.", err))
			return err
		}

		err = t.GetPermissionInfo().SetRole(r)
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not set role.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors("Set role for ", t.GetUsername(), " to ", r.String()))
		return nil
	})
}

func (cm *CommandManager) executeSetRank(r multi.Rank) brigodier.Command {
	return command.Command(func(c *command.Context) error {
		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not set rank.", err))
			return err
		}

		err = t.GetPermissionInfo().SetRank(r)
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not set rank.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors("Set rank for ", t.GetUsername(), " to ", r.String()))
		return nil
	})
}

func (cm *CommandManager) executeGetRole() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not get role.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors("", t.GetUsername(), "'s role is ", t.GetPermissionInfo().GetRole().String()))
		return nil
	})
}

func (cm *CommandManager) executeGetRank() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not get rank.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors("", t.GetUsername(), "'s rank is ", t.GetPermissionInfo().GetRank().String()))
		return nil
	})
}

func (cm *CommandManager) executeIncorrectPermissionCommandUsage() brigodier.Command {
	return command.Command(func(context *command.Context) error {
		context.SendMessage(util.TextWarn("Incorrect usage:\n 1. /permission set role <player> <role>\n 2. /permission set rank <player> <rank>\n 3. /permission get role <player>\n 4. /permission get rank <player>"))
		return nil
	})
}

func (cm *CommandManager) executeIncorrectPermissionCommandSetUsage() brigodier.Command {
	return command.Command(func(context *command.Context) error {
		context.SendMessage(util.TextWarn("Incorrect usage: /permission set role/rank <target> <role/rank>"))
		return nil
	})
}

func (cm *CommandManager) executeIncorrectPermissionCommandGetUsage() brigodier.Command {
	return command.Command(func(context *command.Context) error {
		context.SendMessage(util.TextWarn("Incorrect usage: /permission get <target>"))
		return nil
	})
}
