package listeners

import (
	"time"

	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/playermanager"
	"go.minekube.com/gate/pkg/util/uuid"
)

type ListenerManager struct {
	m   event.Manager
	l   *logger.Logger
	db  *database.Database
	mpm *playermanager.MultiPlayerManager
	id  uuid.UUID
}

func Init(m event.Manager, l *logger.Logger, db *database.Database, mpm *playermanager.MultiPlayerManager, id uuid.UUID) (*ListenerManager, error) {
	now := time.Now()
	lm := &ListenerManager{
		m:   m,
		l:   l,
		db:  db,
		mpm: mpm,
		id:  id,
	}

	err := lm.initFavicon()
	if err != nil {
		return nil, err
	}

	lm.registerListeners()

	lm.l.Info("initialized listener manager", "duration", time.Since(now))
	return lm, nil
}

func (lm *ListenerManager) registerListeners() {
	event.Subscribe(lm.m, 0, lm.onProxyJoin)
	event.Subscribe(lm.m, 0, lm.onServerJoin)
	event.Subscribe(lm.m, 0, lm.onLogin)
	event.Subscribe(lm.m, 0, lm.onDisconnect)
	event.Subscribe(lm.m, 0, lm.onPing)
}
