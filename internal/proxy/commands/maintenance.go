package commands

import (
	"go.minekube.com/brigodier"
)

func (cm *CommandManager) maintenanceCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name)
}
