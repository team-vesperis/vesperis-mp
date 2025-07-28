package multiplayer

import (
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type MultiPlayer struct {
	// The proxy id on which the underlying player is located
	// Can be nil if not online!
	p string

	// The backend id on which the underlying player is located
	// Can be nil if not online!
	b string

	// The id of the underlying player
	id string

	// The username of the underlying player
	name string

	online bool

	friends []*MultiPlayer

	mu sync.RWMutex

	l *logger.Logger
}

// New returns a new MultiPlayer
func New(p proxy.Player, p_id, b_id string, l *logger.Logger) *MultiPlayer {
	mp := &MultiPlayer{
		p:    p_id,
		b:    b_id,
		id:   p.ID().String(),
		name: p.Username(),
		l:    l,
	}

	l.Info("created new multiplayer", "mp", mp, "playerId", p.ID().String())
	return mp
}

func (mp *MultiPlayer) ProxyId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.p
}

func (mp *MultiPlayer) BackendId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.b
}

func (mp *MultiPlayer) SetBackendId(b string) {
	mp.mu.Lock()
	mp.b = b
	mp.mu.Unlock()
}

func (mp *MultiPlayer) Id() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.id
}

func (mp *MultiPlayer) Name() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.name
}

func (mp *MultiPlayer) IsOnline() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.online
}

func (mp *MultiPlayer) SetOnline(online bool) {
	mp.mu.Lock()
	mp.online = online
	mp.mu.Unlock()
}
