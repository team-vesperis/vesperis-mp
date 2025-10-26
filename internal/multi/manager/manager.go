package manager

import (
	"sync"

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

	db *database.Database
	l  *logger.Logger
}

func Init(db *database.Database, l *logger.Logger) *MultiManager {
	mm := &MultiManager{
		proxyMap:  make(map[uuid.UUID]*multi.Proxy),
		playerMap: map[uuid.UUID]*multi.Player{},
		db:        db,
		l:         l,
	}

	multi.SetMultiManager(mm)

	return mm
}

func (mm *MultiManager) GetOwnerMultiProxy() *multi.Proxy {
	return mm.ownerMP
}
