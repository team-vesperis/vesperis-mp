package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
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

func (cm CommandManager) registerCommands() {
	cm.m.Register(cm.databaseCommand("database"))
	cm.m.Register(cm.databaseCommand("db"))
}
