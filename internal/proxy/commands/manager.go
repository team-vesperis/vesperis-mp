package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"github.com/team-vesperis/vesperis-mp/internal/multiproxy/util"
	"go.minekube.com/brigodier"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type CommandManager struct {
	m   *command.Manager
	l   *logger.Logger
	db  *database.Database
	mpm *multiplayer.MultiPlayerManager
}

func Init(p *proxy.Proxy, l *logger.Logger, db *database.Database, mpm *multiplayer.MultiPlayerManager) *CommandManager {
	cm := &CommandManager{
		m:   p.Command(),
		l:   l,
		db:  db,
		mpm: mpm,
	}

	cm.registerCommands()
	return cm
}

var (
	ComponentOnlyPlayersCommand = &component.Text{
		Content: "Only players can use this command.",
		S: component.Style{
			Color: color.Red,
		},
	}

	ErrOnlyPlayersCommand = errors.New("only players can use this command")

	ComponentOnlyPlayersSubCommand = &component.Text{
		Content: "Only players can use this sub command.",
		S: component.Style{
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

func (cm *CommandManager) getMultiPlayerFromTarget(t string, c *command.Context) (*multiplayer.MultiPlayer, error) {
	// target can be a player name or an uuid
	id, err := uuid.Parse(t)
	if err != nil {
		// try to get the id from a player name
		l, err := cm.mpm.GetAllMultiPlayers()
		if err != nil {
			c.SendMessage(&component.Text{
				Content: "Could not get vanish.",
				S: component.Style{
					Color:      util.ColorRed,
					HoverEvent: component.ShowText(&component.Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
				},
			})
			return nil, err
		}

		id = uuid.Nil
		for _, mp := range l {
			if t == mp.GetName() {
				id = mp.GetId()
				break
			}
		}

		if id == uuid.Nil {
			c.SendMessage(&component.Text{
				Content: "Target not found.",
				S:       util.StyleColorOrange,
			})
			return nil, ErrTargetNotFound
		}
	}

	mp, err := cm.mpm.GetMultiPlayer(id)
	if err != nil {
		c.SendMessage(&component.Text{
			Content: "Could not get vanish.",
			S: component.Style{
				Color:      util.ColorRed,
				HoverEvent: component.ShowText(&component.Text{Content: "Internal error: " + err.Error(), S: util.StyleColorRed}),
			},
		})
		return nil, err
	}

	return mp, nil
}

func (cm *CommandManager) SuggestAllMultiPlayers() brigodier.SuggestionProvider {
	return command.SuggestFunc(func(c *command.Context, b *brigodier.SuggestionsBuilder) *brigodier.Suggestions {
		// dynamically update suggestions based on which players are suggested
		//r := b.Remaining

		return b.Build()
	})
}

func (cm *CommandManager) SuggestAllOnlineMultiPlayers() {

}
