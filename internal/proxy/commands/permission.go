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
					Executes(cm.executeIncorrectUsage("/permission set role <target> <role>")).
					Suggests(cm.suggestAllMultiPlayers(false, false))).
				Executes(cm.executeIncorrectUsage("/permission set role <target> <role>"))).
			Executes(cm.executeIncorrectUsage("/permission set role/rank <target> <role/rank>")).
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
					Executes(cm.executeIncorrectUsage("/permission set rank <target> <rank>")).
					Suggests(cm.suggestAllMultiPlayers(false, false))).
				Executes(cm.executeIncorrectUsage("/permission set rank <target> <rank>"))).
			Requires(cm.requireAdmin())).
		Then(brigodier.Literal("get").
			Then(brigodier.Literal("rank").
				Executes(cm.executeIncorrectUsage("/permission get rank <target>")).
				Then(brigodier.Argument("target", brigodier.SingleWord).
					Executes(cm.executeGetRank()).
					Suggests(cm.suggestAllMultiPlayers(false, false)))).
			Then(brigodier.Literal("role").
				Executes(cm.executeIncorrectUsage("/permission get role <target>")).
				Then(brigodier.Argument("target", brigodier.SingleWord).
					Executes(cm.executeGetRole()).
					Suggests(cm.suggestAllMultiPlayers(false, false)))).
			Executes(cm.executeIncorrectUsage("/permission get role/rank <target>"))).
		Executes(cm.executeIncorrectUsage("\n 1. /permission set role/rank <target> <role/rank>\n 2. /permission get role/rank <target>")).
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

		c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorLightBlue), "Set role for ", t.GetUsername(), " to ", r.String()))
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

		c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorLightBlue), "Set rank for ", t.GetUsername(), " to ", r.String()))
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

		c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightBlue, util.ColorLightGreen), t.GetUsername(), "'s role is ", t.GetPermissionInfo().GetRole().String()))
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

		c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightBlue, util.ColorLightGreen), t.GetUsername(), "'s rank is ", t.GetPermissionInfo().GetRank().String()))
		return nil
	})
}
