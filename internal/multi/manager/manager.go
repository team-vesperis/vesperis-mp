package manager

import (
	"sync"
	"time"

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

func Init(cf *config.Config, db *database.Database, l *logger.Logger) (*MultiManager, error) {
	now := time.Now()

	mm := &MultiManager{
		proxyMap:   map[uuid.UUID]*multi.Proxy{},
		playerMap:  map[uuid.UUID]*multi.Player{},
		backendMap: map[uuid.UUID]*multi.Backend{},
		cf:         cf,
		db:         db,
		l:          l,
	}

	multi.SetMultiManager(mm)

	id, err := mm.CreateNewProxyId()
	if err != nil {
		return &MultiManager{}, err
	}

	mm.l.Debug("Found a id to use for this proxy.", "id", id)

	_, err = mm.NewMultiProxy(id)
	if err != nil {
		return &MultiManager{}, err
	}

	// start update listeners
	mm.db.CreateListener(multi.UpdateMultiPlayerChannel, mm.createUpdateListener())
	mm.db.CreateListener(multi.UpdateMultiBackendChannel, mm.createBackendUpdateListener())
	mm.db.CreateListener(multi.UpdateMultiProxyChannel, mm.createProxyUpdateListener())

	_, err = mm.GetAllMultiProxiesFromDatabase()
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

	mm.l.Info("initialized multimanager", "duration", time.Since(now))
	return mm, nil
}

func (mm *MultiManager) Close() error {
	now := time.Now()

	l := mm.ownerMP.GetBackendsIds()
	for _, id := range l {
		err := mm.DeleteMultiBackend(id)
		if err != nil {
			return err
		}
	}

	err := mm.DeleteMultiProxy(mm.ownerMP.GetId())
	if err != nil {
		return err
	}

	mm.hbm.stop()

	mm.l.Info("multimanager closed successfully", "duration", time.Since(now))
	return nil
}

func (mm *MultiManager) GetOwnerMultiProxy() *multi.Proxy {
	return mm.ownerMP
}
