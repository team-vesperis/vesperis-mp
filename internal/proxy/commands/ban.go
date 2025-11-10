package commands

import "go.minekube.com/brigodier"

func (cm *CommandManager) banCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name)
}
