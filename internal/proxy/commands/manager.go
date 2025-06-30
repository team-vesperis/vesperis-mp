package commands

import (
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"go.minekube.com/gate/pkg/command"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

type CommandManager struct {
	m  *command.Manager
	l  *zap.SugaredLogger
	db database.Database
}

func Init(p *proxy.Proxy, l *zap.SugaredLogger, db database.Database) CommandManager {
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
