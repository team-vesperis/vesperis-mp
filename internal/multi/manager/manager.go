package manager

import (
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiManager struct {
	proxyMap   map[uuid.UUID]*multi.Proxy
	playerMap  map[uuid.UUID]*multi.Player
	backendMap map[uuid.UUID]*multi.Backend
	mu         sync.RWMutex

	ownerMP *multi.Proxy

	hbm *hartBeatManager

	cf *config.Config
	db *database.Database
	l  *logger.Logger
}

func Init(cf *config.Config, db *database.Database, l *logger.Logger) *MultiManager {
	mm := &MultiManager{
		proxyMap:   map[uuid.UUID]*multi.Proxy{},
		playerMap:  map[uuid.UUID]*multi.Player{},
		backendMap: map[uuid.UUID]*multi.Backend{},
		cf:         cf,
		db:         db,
		l:          l,
	}

	multi.SetMultiManager(mm)

	return mm
}

func (mm *MultiManager) Start() {
	// start update listeners
	mm.db.CreateListener(multi.UpdateMultiPlayerChannel, mm.createUpdateListener())
	mm.db.CreateListener(multi.UpdateMultiBackendChannel, mm.createBackendUpdateListener())
	mm.db.CreateListener(multi.UpdateMultiProxyChannel, mm.createProxyUpdateListener())

	_, err := mm.GetAllMultiProxiesFromDatabase()
	if err != nil {
		mm.l.Warn("filling up multiproxy map error", "error", err)
	}

	_, err = mm.GetAllMultiBackendsFromDatabase()
	if err != nil {
		mm.l.Warn("filling up multibackend map error", "error", err)
	}

	_, err = mm.GetAllMultiPlayersFromDatabase()
	if err != nil {
		mm.l.Warn("filling up multiplayer map error", "error", err)
	}
}

func (mm *MultiManager) Close() error {
	l := mm.ownerMP.GetBackendsIds()
	mm.l.Debug("deleting list", "list", l)
	for _, id := range l {
		err := mm.DeleteMultiBackend(id)
		if err != nil {
			return err
		}

		mm.l.Debug("deleted backend")
	}

	err := mm.DeleteMultiProxy(mm.ownerMP.GetId())
	if err != nil {
		return err
	}

	mm.hbm.stop()

	mm.l.Debug("multimanager closed successfully")
	return nil
}

func (mm *MultiManager) GetOwnerMultiProxy() *multi.Proxy {
	return mm.ownerMP
}
