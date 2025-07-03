package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type CommandManager struct {
	m  *command.Manager
	l  *logger.Logger
	db *database.Database
}

func Init(p *proxy.Proxy, l *logger.Logger, db *database.Database) CommandManager {
	cm := CommandManager{
		m:  p.Command(),
		l:  l,
		db: db,
	}

	cm.registerCommands()
	return cm
}

func (cm CommandManager) registerCommands() error {
	cm.m.Register(cm.databaseCommand("database"))
	cm.m.Register(cm.databaseCommand("db"))
	return nil
}
