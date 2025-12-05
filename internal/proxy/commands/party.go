package commands

import "go.minekube.com/brigodier"

func (cm *CommandManager) partyCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name)
}
