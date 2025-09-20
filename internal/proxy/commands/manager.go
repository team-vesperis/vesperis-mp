package commands

import (
	"errors"
	"strings"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/playermanager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	. "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type CommandManager struct {
	m   *command.Manager
	l   *logger.Logger
	db  *database.Database
	mpm *playermanager.MultiPlayerManager
}

func Init(p *proxy.Proxy, l *logger.Logger, db *database.Database, mpm *playermanager.MultiPlayerManager) (*CommandManager, error) {
	cm := &CommandManager{
		m:   p.Command(),
		l:   l,
		db:  db,
		mpm: mpm,
	}

	cm.registerCommands()
	return cm, nil
}

var (
	ComponentOnlyPlayersCommand = &Text{
		Content: "Only players can use this command.",
		S: Style{
			Color: color.Red,
		},
	}

	ErrOnlyPlayersCommand = errors.New("only players can use this command")

	ComponentOnlyPlayersSubCommand = &Text{
		Content: "Only players can use this sub command.",
		S: Style{
			Color: color.Red,
		},
	}

	ErrOnlyPlayersSubCommand = errors.New("only players can use this sub command")
)

func (cm *CommandManager) registerCommands() {
	cm.m.Register(cm.databaseCommand("database"))
	cm.m.Register(cm.databaseCommand("db"))
	cm.m.Register(cm.vanishCommand("vanish"))
	cm.m.Register(cm.vanishCommand("v"))
}

var ErrTargetNotFound = errors.New("target not found")
var ComponentTargetNotFound = &Text{
	Content: "Target not found.",
	S:       util.StyleColorOrange,
}

func (cm *CommandManager) getMultiPlayerFromTarget(t string, c *command.Context) (*multi.MultiPlayer, error) {
	// target can be a player name or an uuid
	id, err := uuid.Parse(t)
	if err != nil {
		// try to get the id from a player name
		l := cm.mpm.GetAllMultiPlayers()

		id = uuid.Nil
		for _, mp := range l {
			if t == mp.GetName() {
				id = mp.GetId()
				break
			}
		}

		if id == uuid.Nil {
			c.SendMessage(ComponentTargetNotFound)
			return nil, ErrTargetNotFound
		}
	}

	mp, err := cm.mpm.GetMultiPlayer(id)
	if err != nil {
		if err == database.ErrDataNotFound {
			c.SendMessage(ComponentTargetNotFound)
			return nil, ErrTargetNotFound
		}

		c.SendMessage(&Text{
			Content: "Could not get vanish.",
			S: Style{
				Color:      util.ColorRed,
				HoverEvent: ShowText(&Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
			},
		})
		return nil, err
	}

	return mp, nil
}

// suggests all multiplayers, online and offline.
// vanished players are hidden from non-privileged players
func (cm *CommandManager) SuggestAllMultiPlayers(onlyOnline bool) brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		r := b.RemainingLowerCase

		if len(r) < 1 {
			b.Suggest("type a playerName or UUID")
			return b.Build()
		}

		// use list to get all names and ids
		var l []*multi.MultiPlayer
		if onlyOnline {
			l = cm.mpm.GetAllOnlinePlayers()
		} else {
			l = cm.mpm.GetAllMultiPlayers()
		}

		hide_vanished := false
		p, ok := c.Source.(proxy.Player)
		if ok {
			mp, err := cm.mpm.GetMultiPlayer(p.ID())
			if err != nil {
				cm.l.Error("suggest all multiplayers get multiplayer error", "error", err)
				return b.Build()
			}

			if !mp.GetPermissionInfo().IsPrivileged() {
				hide_vanished = true
			}
		}

		for _, mp := range l {
			if hide_vanished && mp.IsVanished() {
				continue
			}

			name := mp.GetName()
			if strings.HasPrefix(strings.ToLower(name), r) {
				b.Suggest(name)
			}

			id := mp.GetId().String()
			if len(r) > 1 && strings.HasPrefix(strings.ToLower(id), r) {
				b.Suggest(id)
			}
		}

		return b.Build()
	})
}
