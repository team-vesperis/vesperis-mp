package commands

import (
	"errors"
	"strings"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
)

func (cm *CommandManager) friendsCommand(name string) brigodier.LiteralNodeBuilder {
	return brigodier.Literal(name).
		Then(brigodier.Literal("decline").
			Executes(cm.executeIncorrectUsage("/friends decline <target>")).
			Then(brigodier.Argument("target", brigodier.SingleWord).
				Suggests(cm.suggestAllFriendRequestMultiPlayers()).
				Executes(cm.executeFriendsResponse(false)))).
		Then(brigodier.Literal("accept").
			Executes(cm.executeIncorrectUsage("/friends accept <target>")).
			Then(brigodier.Argument("target", brigodier.SingleWord).
				Suggests(cm.suggestAllFriendRequestMultiPlayers()).
				Executes(cm.executeFriendsResponse(true)))).
		Then(brigodier.Literal("request").
			Executes(cm.executeIncorrectUsage("/friends request <target>")).
			Then(brigodier.Argument("target", brigodier.SingleWord).
				Suggests(cm.suggestAllNonFriendMultiPlayers()).
				Executes(cm.executeFriendsRequest()))).
		Then(brigodier.Literal("list").
			Executes(cm.executeFriendsList())).
		Then(brigodier.Literal("remove").
			Executes(cm.executeIncorrectUsage("/friends remove <target>")).
			Then(brigodier.Argument("target", brigodier.SingleWord).
				Suggests(cm.suggestAllFriends()).
				Executes(cm.executeFriendsRemove()))).
		Executes(cm.executeIncorrectUsage("\n 1. /friends request <target>\n 2. /friends list\n 3. /friends accept <target>\n 4. /friends decline <target>"))
}

func (cm *CommandManager) executeFriendsRemove() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			p.SendMessage(util.TextInternalError("Could not remove friend.", err))
			return err
		}

		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}

			p.SendMessage(util.TextInternalError("Could not remove friend.", err))
			return err
		}

		if mp.GetFriendInfo().IsFriend(t.GetId()) && t.GetFriendInfo().IsFriend(mp.GetId()) {
			err := mp.GetFriendInfo().RemoveFriendId(t.GetId())
			if err != nil {
				p.SendMessage(util.TextInternalError("Could not remove friend.", err))
				return err
			}

			err = t.GetFriendInfo().RemoveFriendId(mp.GetId())
			if err != nil {
				p.SendMessage(util.TextInternalError("Could not remove friend.", err))
				return err
			}

			return nil
		}

		// target is not requesting friend request
		p.SendMessage(util.TextWarn("Target is not your friend."))
		return nil
	})
}

func (cm *CommandManager) executeFriendsResponse(accept bool) brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			p.SendMessage(util.TextInternalError("Could not respond to friend request.", err))
			return err
		}

		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}

			p.SendMessage(util.TextInternalError("Could not respond to friend request.", err))
			return err
		}

		if mp.GetFriendInfo().IsFriendRequest(t.GetId()) && t.GetFriendInfo().IsPendingFriend(mp.GetId()) {
			err := mp.GetFriendInfo().RemoveFriendRequestId(t.GetId())
			if err != nil {
				p.SendMessage(util.TextInternalError("Could not respond to friend request.", err))
				return err
			}

			err = t.GetFriendInfo().RemovePendingFriendRequestId(mp.GetId())
			if err != nil {
				p.SendMessage(util.TextInternalError("Could not respond to friend request.", err))
				return err
			}

			if accept {
				err = mp.GetFriendInfo().AddFriendId(t.GetId())
				if err != nil {
					p.SendMessage(util.TextInternalError("Could not respond to friend request.", err))
					return err
				}

				err = t.GetFriendInfo().AddFriendId(mp.GetId())
				if err != nil {
					p.SendMessage(util.TextInternalError("Could not respond to friend request.", err))
					return err
				}
			}

			if accept {
				p.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorLightBlue), "You have accepted ", t.GetUsername(), "'s friend request."))
			} else {
				p.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorOrange, util.ColorLightBlue), "You have ", "declined ", t.GetUsername(), "'s friend request."))
			}

			if t.IsOnline() {
				mproxy := t.GetProxy()
				if mproxy == nil {
					return nil
				}

				var tr *task.TaskResponse
				if accept {
					tr = cm.tm.BuildTask(tasks.NewMessageTask(t.GetId(), mproxy.GetId(), util.ComponentToString(util.TextAlternatingColors(util.ColorList(util.ColorLightBlue, util.ColorLightGreen), mp.GetUsername(), " has accepted your friend request!"))))
				} else {
					tr = cm.tm.BuildTask(tasks.NewMessageTask(t.GetId(), mproxy.GetId(), util.ComponentToString(util.TextAlternatingColors(util.ColorList(util.ColorLightBlue, util.ColorOrange), mp.GetUsername(), " has declined your friend request."))))
				}

				if !tr.IsSuccessful() {
					err := errors.New(tr.GetInfo())
					p.SendMessage(util.TextInternalError("Could not respond to friend request.", err))
					return err
				}
			}

			return nil
		}

		// target is not requesting friend request
		p.SendMessage(util.TextWarn("Target did not request friend request."))
		return nil
	})
}

func (cm *CommandManager) executeFriendsRequest() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not send friends request.", err))
			return err
		}

		t, err := cm.getMultiPlayerFromTarget(c.String("target"))
		if err != nil {
			if err == ErrTargetNotFound {
				c.SendMessage(TextTargetNotFound)
				return nil
			}

			c.SendMessage(util.TextInternalError("Could not send friends request.", err))
			return err
		}

		err = t.GetFriendInfo().AddFriendRequestId(mp.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not send friends request.", err))
			return err
		}

		err = mp.GetFriendInfo().AddPendingFriendRequestId(t.GetId())
		if err != nil {
			c.SendMessage(util.TextInternalError("Could not send friends request.", err))
			return err
		}

		// send message to target
		if t.IsOnline() {
			mproxy := t.GetProxy()
			if mproxy == nil {
				c.SendMessage(util.TextInternalError("Could not send friends request.", multi.ErrProxyNilWhileOnline))
				return multi.ErrProxyNilWhileOnline
			}

			tr := cm.tm.BuildTask(tasks.NewMessageTask(t.GetId(), mproxy.GetId(), util.ComponentToString(&component.Text{
				Content: "Received friend request from ",
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
								Content: "Click to accept friend request.",
								S:       util.StyleColorGray,
							}),
							ClickEvent: component.SuggestCommand("/friends accept " + mp.GetUsername()),
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
								Content: "Click to decline friend request.",
								S:       util.StyleColorGray,
							}),
							ClickEvent: component.SuggestCommand("/friends decline " + mp.GetUsername()),
						},
					},
				},
			})))
			if !tr.IsSuccessful() {
				err := errors.New(tr.GetInfo())
				c.SendMessage(util.TextInternalError("Could not send friends request.", err))
				return err
			}
		}

		p.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorLightGreen, util.ColorLightBlue), "Successfully send friend request to ", t.GetUsername(), "."))

		return nil
	})
}

func (cm *CommandManager) executeFriendsList() brigodier.Command {
	return command.Command(func(c *command.Context) error {
		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			c.SendMessage(ComponentOnlyPlayersSubCommand)
			return ErrOnlyPlayersSubCommand
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			return err
		}

		if len(mp.GetFriendInfo().GetFriendsIds()) < 1 {

			return nil
		}

		c.SendMessage(util.TextSuccessful("Your friend list:"))
		for _, id := range mp.GetFriendInfo().GetFriendsIds() {
			f, err := cm.mm.GetMultiPlayer(id)
			if err != nil {
				continue
			}

			var hover component.Component
			if f.IsOnline() {
				hover = util.TextSuccessful("Player is online!")
			} else {
				hover = util.TextAlternatingColors(util.ColorList(util.ColorOrange, util.ColorLightGreen, util.ColorLightBlue), "Player is offline. ", "Last seen: ", util.FormatTimeSince(*f.GetLastSeen()), "", " ago.")
			}

			c.SendMessage(&component.Text{
				Content: " - ",
				S:       util.StyleColorLightGreen,
				Extra: []component.Component{
					&component.Text{
						Content: f.GetUsername(),
						S: component.Style{
							Color:      util.ColorLightBlue,
							HoverEvent: component.ShowText(hover),
						},
					},
				},
			})
		}

		return nil
	})
}

func (cm *CommandManager) suggestAllNonFriendMultiPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase

		if len(r) < 1 {
			b.Suggest("type a username or UUID...")
			return b.Build()
		}

		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			return b.Build()
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			cm.l.Error("suggest all non friend multiplayers get multiplayer error", "error", err)
			return b.Build()
		}

		for _, t := range cm.mm.GetAllMultiPlayers(mp.GetPermissionInfo().IsPrivileged()) {
			if mp.GetFriendInfo().IsFriend(t.GetId()) && t.GetFriendInfo().IsFriend(mp.GetId()) {
				continue
			}

			username := t.GetUsername()
			if strings.HasPrefix(strings.ToLower(username), r) {
				b.Suggest(username)
			}

			id := t.GetId().String()
			if len(r) > 2 && strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}

func (cm *CommandManager) suggestAllFriends() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase

		if len(r) < 1 {
			b.Suggest("type a username or UUID...")
			return b.Build()
		}

		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			return b.Build()
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			cm.l.Error("suggest all friends get multiplayer error", "error", err)
			return b.Build()
		}

		for _, id := range mp.GetFriendInfo().GetFriendsIds() {
			t, err := cm.mm.GetMultiPlayer(id)
			if err != nil {
				continue
			}

			username := t.GetUsername()
			if strings.HasPrefix(strings.ToLower(username), r) {
				b.Suggest(username)
			}

			id := t.GetId().String()
			if len(r) > 2 && strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}

func (cm *CommandManager) suggestAllFriendRequestMultiPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase

		if len(r) < 1 {
			b.Suggest("type a username or UUID...")
			return b.Build()
		}

		p := cm.getGatePlayerFromSource(c.Source)
		if p == nil {
			return b.Build()
		}

		mp, err := cm.mm.GetMultiPlayer(p.ID())
		if err != nil {
			cm.l.Error("suggest all friend requests get multiplayer error", "error", err)
			return b.Build()
		}

		for _, id := range mp.GetFriendInfo().GetFriendRequestIds() {
			t, err := cm.mm.GetMultiPlayer(id)
			if err != nil {
				continue
			}

			username := t.GetUsername()
			if strings.HasPrefix(strings.ToLower(username), r) {
				b.Suggest(username)
			}

			id := t.GetId().String()
			if len(r) > 2 && strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}
