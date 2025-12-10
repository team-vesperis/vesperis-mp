package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (cm *CommandManager) partyCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("decline").
			Then(brigodier.Argument("partyId", brigodier.SingleWord).
				Executes(cm.executePartyResponse(false)))).
		Then(brigodier.Literal("accept").
			Then(brigodier.Argument("partyId", brigodier.SingleWord).
				Executes(cm.executePartyResponse(true)))).
		Then(brigodier.Literal("invite").
			Then(brigodier.Argument("target", brigodier.SingleWord).
				Suggests(cm.suggestAllMultiPlayers(true, true)).
				Executes(cm.executePartyInvite()))).
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

		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}

			c.SendMessage(util.TextInternalError("Could not invite to party.", err))
			return err
		}

		if t.GetId() == p.ID() {
			c.SendMessage(util.TextWarn("You can't invite yourself to a party."))
			return nil
		}

		if !t.IsOnline() {
			c.SendMessage(TextTargetIsOffline)
			return nil
		}

		proxy := t.GetProxy()
		if proxy == nil {
			c.SendMessage(util.TextInternalError("Could not invite to party.", multi.ErrProxyNilWhileOnline))
			return multi.ErrProxyNilWhileOnline
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not invite to party.", err))
			return err
		}

		partyId := mp.GetPartyId()
		var party *multi.Party
		if partyId == uuid.Nil {
			party, err = cm.mm.NewMultiParty(mp.GetId())
			if err != nil {
				c.SendMessage(util.TextInternalError("Could not invite to party.", err))
				return err
			}
		} else {
			party, err = cm.mm.GetMultiParty(partyId)
			if err != nil {
				c.SendMessage(util.TextInternalError("Could not invite to party.", err))
				return err
			}
		}

		if t.IsInvitedToParty(party.GetId()) {
			c.SendMessage(util.TextWarn("Target is already invited to the party."))
			return nil
		}

		err = party.AddPartyInvitation(t.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not invite to party.", err))
			return err
		}

		err = t.AddPartyInvitation(party.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not invite to party.", err))
			return err
		}

		tr := cm.tm.BuildTask(tasks.NewMessageTask(t.GetId(), proxy.GetId(), util.ComponentToString(&component.Text{
			Content: "Received party invitation from ",
			S:       util.StyleColorLightGreen,
			Extra: []component.Component{
				&component.Text{
					Content: mp.GetUsername(),
					S:       util.StyleColorLightBlue,
				},
				&component.Text{
					Content: ". ",
					S:       util.StyleColorLightGreen,
				},
				&component.Text{
					Content: "[ACCEPT]",
					S: component.Style{
						Color: util.ColorGreen,
						Bold:  component.True,
						HoverEvent: component.ShowText(&component.Text{
							Content: "Click to accept party invitation.",
							S:       util.StyleColorGray,
						}),
						ClickEvent: component.SuggestCommand("/party accept " + party.GetId().String()),
					},
				},
				&component.Text{
					Content: " - ",
					S:       util.StyleColorGray,
				},
				&component.Text{
					Content: "[DECLINE]",
					S: component.Style{
						Color: util.ColorRed,
						Bold:  component.True,
						HoverEvent: component.ShowText(&component.Text{
							Content: "Click to decline party invitation.",
							S:       util.StyleColorGray,
						}),
						ClickEvent: component.SuggestCommand("/party decline " + party.GetId().String()),
					},
				},
			},
		})))

		if !tr.IsSuccessful() {
			err := errors.New(tr.GetInfo())
			c.SendMessage(util.TextInternalError("Could not invite to party.", err))
			return err
		}

		c.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorLightBlue), "Successfully send party invitation to ", t.GetUsername()))
		return nil
	})
}

func (cm *CommandManager) executePartyResponse(accept bool) brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not respond to party invite.", err))
			return err
		}

		partyId, err := uuid.Parse(c.String("partyId"))
		if err != nil {
			c.SendMessage(util.TextWarn("Invalid Party UUID"))
			return nil
		}

		party, err := cm.mm.GetMultiParty(partyId)
		if err != nil {
			if err == database.ErrDataNotFound {
				c.SendMessage(util.TextWarn("Party doesn't exist."))
				return nil
			}

			c.SendMessage(util.TextInternalError("Could not respond to party invite.", err))
			return err
		}

		err = party.RemovePartyInvitation(mp.GetId())
		if err != nil {
			if err == multi.ErrPlayerNotFound {
				c.SendMessage(util.TextWarn("The specified party did not invite you."))
				return nil
			}

			c.SendMessage(util.TextInternalError("Could not respond to party invite.", err))
			return err
		}

		err = mp.RemovePartyInvitation(party.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not respond to party invite.", err))
			return err
		}

		if accept {
			_, err = cm.leaveParty(mp, c)
			if err != nil {
				c.SendMessage(util.TextInternalError("Could not accept party invite.", err))
				return err
			}

			err = party.AddPartyMember(mp.GetId())
			if err != nil {
				if err == multi.ErrPlayerNotFound {
					c.SendMessage(util.TextWarn("The specified party did not invite you."))
					return nil
				}

				c.SendMessage(util.TextInternalError("Could not accept party invite.", err))
				return err
			}

			err = mp.SetPartyId(party.GetId())
			if err != nil {
				c.SendMessage(util.TextInternalError("Could not accept party invite.", err))
				return err
			}

			for _, id := range party.GetPartyMembers() {
				member, err := cm.mm.GetMultiPlayer(id)
				if err != nil {
					return err
				}

				if member.IsOnline() {
					proxy := member.GetProxy()
					if proxy == nil {
						cm.l.Warn("party accept command member online but proxy nil error", "memberId", member.GetId(), "error", multi.ErrProxyNilWhileOnline)
						continue
					}

					tr := cm.tm.BuildTask(tasks.NewMessageTask(member.GetId(), proxy.GetId(), util.ComponentToString(util.TextAlternatingColors(util.ColorList(util.ColorOrange, util.ColorLightBlue, util.ColorGray), "[Party]: ", mp.GetUsername(), " has joined the party."))))
					if !tr.IsSuccessful() {
						cm.l.Warn("party accept command send message to other members error", "memberId", member.GetId(), "error", tr.GetInfo())
					}
				}
			}

			c.SendMessage(util.TextSuccessful("Party invite accepted."))

		} else {
			c.SendMessage(util.TextSuccessful("Party invite declined."))
		}
		return nil
	})
}

func (cm *CommandManager) executePartyLeave() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not leave party.", err))
			return err
		}

		wasNotInParty, err := cm.leaveParty(mp, c)
		if wasNotInParty {
			c.SendMessage(util.TextWarn("You are not in a party."))
			return nil
		}

		if err != nil {
			c.SendMessage(util.TextInternalError("Could not leave party.", err))
			return err
		}

		c.SendMessage(util.TextSuccessful("Successfully left the party."))
		return nil
	})
}

func (cm *CommandManager) leaveParty(mp *multi.Player, c *command.Context) (bool, error) {
	id := mp.GetPartyId()
	if id == uuid.Nil {
		return true, nil
	}

	party, err := cm.mm.GetMultiParty(id)
	if err != nil {
		return false, err
	}

	err = mp.SetPartyId(uuid.Nil)
	if err != nil {
		c.SendMessage(util.TextInternalError("Could not leave party.", err))
		return false, err
	}

	err = party.RemovePartyMember(mp.GetId())
	if err != nil {
		c.SendMessage(util.TextInternalError("Could not leave party.", err))
		return false, err
	}

	// last player in party
	if len(party.GetPartyMembers()) < 1 {
		err = cm.mm.DeleteMultiParty(party.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not leave party.", err))
			return false, err
		}
	} else {
		if party.GetPartyOwner() == mp.GetId() {
			err = party.SetPartyOwner(party.GetPartyMembers()[0])
			if err != nil {
				c.SendMessage(util.TextInternalError("Could not leave party.", err))
				return false, err
			}
		}

		for _, id := range party.GetPartyMembers() {
			member, err := cm.mm.GetMultiPlayer(id)
			if err != nil {
				return false, err
			}

			if member.IsOnline() {
				proxy := member.GetProxy()
				if proxy == nil {
					cm.l.Warn("party leave command member online but proxy nil error", "memberId", member.GetId(), "error", multi.ErrProxyNilWhileOnline)
					continue
				}

				tr := cm.tm.BuildTask(tasks.NewMessageTask(member.GetId(), proxy.GetId(), util.ComponentToString(util.TextAlternatingColors(util.ColorList(util.ColorOrange, util.ColorLightBlue, util.ColorGray), "[Party]: ", mp.GetUsername(), " has left the party."))))
				if !tr.IsSuccessful() {
					cm.l.Warn("party leave command send message to other members error", "memberId", member.GetId(), "error", tr.GetInfo())
				}
			}
		}
	}

	return false, nil
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

		party, err := cm.mm.NewMultiParty(mp.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not create party.", err))
			return err
		}

		err = mp.SetPartyId(party.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not create party.", err))
			return err
		}

		c.SendMessage(util.TextSuccessful("Successfully created a party."))
		return nil
	})
}
