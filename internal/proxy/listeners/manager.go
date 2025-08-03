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
	pdm *multiplayer.PlayerDataManager
	mpm *multiplayer.MultiPlayerManager
}

func Init(m event.Manager, l *logger.Logger, db *database.Database, pdm *multiplayer.PlayerDataManager, mpm *multiplayer.MultiPlayerManager) *ListenerManager {
	lm := &ListenerManager{
		m:   m,
		l:   l,
		db:  db,
		pdm: pdm,
		mpm: mpm,
	}

	lm.registerListeners()
	return lm
}

func (lm *ListenerManager) registerListeners() {
	event.Subscribe(lm.m, 0, lm.onJoin)
}
