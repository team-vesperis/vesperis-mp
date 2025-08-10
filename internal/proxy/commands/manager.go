package commands

import (
	"errors"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
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

func (cm CommandManager) registerCommands() {
	cm.m.Register(cm.databaseCommand("database"))
	cm.m.Register(cm.databaseCommand("db"))
	cm.m.Register(cm.vanishCommand("vanish"))
	cm.m.Register(cm.vanishCommand("v"))
}
