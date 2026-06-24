package commands

import (
	"go.minekube.com/brigodier"
)

func (cm *CommandManager) maintenanceCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("status")).
		Then(brigodier.Literal("list")).
		Then(brigodier.Literal("activate")).
		Then(brigodier.Literal("deactivate"))
}
