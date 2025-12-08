package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (cm *CommandManager) partyCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("leave").
			Executes(cm.executePartyLeave())).
		Then(brigodier.Literal("create").
			Executes(cm.executePartyCreate()))
}

func (cm *CommandManager) executePartyInvite() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		// mp, err := cm.mm.GetMultiPlayer(p.ID())
		// if err != nil {
		// 	c.SendMessage(util.TextInternalError("Could not invite to party.", err))
		// 	return err
		// }

		return nil
	})
}

func (cm *CommandManager) executePartyAccept() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		// mp, err := cm.mm.GetMultiPlayer(p.ID())
		// if err != nil {
		// 	c.SendMessage(util.TextInternalError("Could not accept party invite.", err))
		// 	return err
		// }

		c.SendMessage(util.TextSuccessful("Party invite accepted."))
		return nil
	})
}

func (cm *CommandManager) executePartyDecline() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		// mp, err := cm.mm.GetMultiPlayer(p.ID())
		// if err != nil {
		// 	c.SendMessage(util.TextInternalError("Could not decline party invite.", err))
		// 	return err
		// }

		// partyId, err := uuid.Parse(c.String("partyId"))
		// if err != nil {
		// 	c.SendMessage(util.TextWarn("Parsed invalid UUID"))
		// 	return nil
		// }

		c.SendMessage(util.TextSuccessful("Party invite declined."))
		return nil
	})
}

func (cm *CommandManager) executePartyLeave() brigodier.Command {
	return command.Command(func(c *command.Context) error {

		return nil
	})
}

func (cm *CommandManager) executePartyCreate() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not create party.", err))
			return err
		}

		if mp.GetPartyId() != uuid.Nil {
			c.SendMessage(util.TextWarn("You are already in a party."))
			return nil
		}

		_, err = cm.mm.NewMultiParty(mp.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not create party.", err))
			return err
		}

		c.SendMessage(util.TextSuccessful("Successfully created a party."))
		return nil
	})
}
