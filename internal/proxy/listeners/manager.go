package listeners

import (
	"time"

	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/manager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type ListenerManager struct {
	m         event.Manager
	l         *logger.Logger
	db        *database.Database
	mm        *manager.MultiManager
	ownerGate *proxy.Proxy
	tm        *task.TaskManager
}

func Init(m event.Manager, l *logger.Logger, db *database.Database, mm *manager.MultiManager, ownerGate *proxy.Proxy, tm *task.TaskManager) (*ListenerManager, error) {
	now := time.Now()
	lm := &ListenerManager{
		m:         m,
		l:         l,
		db:        db,
		mm:        mm,
		ownerGate: ownerGate,
		tm:        tm,
	}

	err := lm.initFavicon()
	if err != nil {
		return nil, err
	}

	err = lm.initResourcePack()
	if err != nil {
		return nil, err
	}

	lm.registerListeners()

	lm.l.Info("initialized listener manager", "duration", time.Since(now))
	return lm, nil
}

func (lm *ListenerManager) registerListeners() {
	event.Subscribe(lm.m, 0, lm.onProxyJoin)
	event.Subscribe(lm.m, 5, lm.onServerJoin)
	event.Subscribe(lm.m, 0, lm.onLogin)
	event.Subscribe(lm.m, 5, lm.onDisconnect)
	event.Subscribe(lm.m, 0, lm.onPing)

	event.Subscribe(lm.m, 0, lm.onRegister)
	event.Subscribe(lm.m, 0, lm.onUnRegister)

	event.Subscribe(lm.m, 0, lm.onChooseInitialServer)
	event.Subscribe(lm.m, 0, lm.onPreShutdown)

	event.Subscribe(lm.m, 5, lm.sendResourcePack)

	event.Subscribe(lm.m, 0, lm.onChatMessage)
}
