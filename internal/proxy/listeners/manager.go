package listeners

import (
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
)

type ListenerManager struct {
	m   event.Manager
	l   *logger.Logger
	db  *database.Database
	ppm *multiplayer.PlayerPermissionManager
	mpm *multiplayer.MultiPlayerManager
	id  string
}

func Init(m event.Manager, l *logger.Logger, db *database.Database, ppm *multiplayer.PlayerPermissionManager, mpm *multiplayer.MultiPlayerManager, id string) *ListenerManager {
	lm := &ListenerManager{
		m:   m,
		l:   l,
		db:  db,
		ppm: ppm,
		mpm: mpm,
		id:  id,
	}

	lm.registerListeners()
	return lm
}

func (lm *ListenerManager) registerListeners() {
	event.Subscribe(lm.m, 0, lm.onProxyJoin)
	event.Subscribe(lm.m, 0, lm.onServerJoin)
	event.Subscribe(lm.m, 0, lm.onLogin)
	event.Subscribe(lm.m, 0, lm.onDisconnect)
}
