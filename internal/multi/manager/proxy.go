package manager

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (mm *MultiManager) createProxyUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		m := msg.Payload
		s := strings.Split(m, "_")
		if len(s) != 3 {
			mm.l.Warn("multiproxy update channel received message with incorrect length", "message", m)
			return
		}

		originProxy := s[0]
		// from own proxy, no update needed
		if mm.ownerMP.GetId().String() == originProxy {
			return
		}

		id, err := uuid.Parse(s[1])
		if err != nil {
			mm.l.Error("multiproxy update channel parse uuid error", "parsed uuid", s[1], "error", err)
			return
		}

		k := s[2]

		mp, err := mm.GetMultiProxy(id)
		if err != nil {
			mm.l.Error("multiproxy update channel get multiproxy error", "proxyId", id, "error", err)
			return
		}

		mm.l.Debug("received proxy update request", "originProxyId", originProxy, "key", k)

		// already created
		if k == "new" {
			return
		}

		if k == "delete" {
			err := mm.deleteMultiProxy(id, false)
			if err != nil {
				mm.l.Error("multiproxy update channel delete multiproxy error", "proxyId", id, "error", err)
			}
			return
		}

		dataKey, err := key.GetProxyKey(k)
		if err != nil {
			mm.l.Error("multiproxy update channel get data key error", "proxyId", id, "key", k, "error", err)
			return
		}

		mp.Update(dataKey)
	}
}

func (mm *MultiManager) NewMultiProxy() (*multi.Proxy, error) {
	now := time.Now()

	var id uuid.UUID
	var err error
	for {
		id = uuid.New()
		_, err = mm.GetMultiBackend(id)
		if err == nil {
			continue
		}

		if err == database.ErrDataNotFound {
			err = nil
		}

		break
	}

	if err != nil {
		return nil, err
	}

	var addr string

	switch mm.cf.GetMode() {
	case config.Mode_Default:
		addr = mm.cf.GetBind()
	case config.Mode_Kubernetes:
		addr = fmt.Sprintf("%s.proxy.default.svc.cluster.local:25565", id.String())
	default:
		return nil, config.ErrIncorrectMode
	}

	data := &data.ProxyData{
		Address:       addr,
		Maintenance:   false,
		Backends:      make([]uuid.UUID, 0),
		Players:       make([]uuid.UUID, 0),
		LastHeartBeat: &now,
	}

	err = mm.db.SetProxyData(id, data)
	if err != nil {
		return nil, err
	}

	mp, err := mm.CreateMultiProxyFromDatabase(id)
	if err != nil {
		return nil, err
	}

	if mm.ownerMP == nil {
		mm.ownerMP = mp
	}

	// update every proxies' map
	m := mm.ownerMP.GetId().String() + "_" + id.String() + "_new"
	err = mm.db.Publish(multi.UpdateMultiProxyChannel, m)
	if err != nil {
		return nil, err
	}

	mm.l.Info("created new multiproxy", "proxyId", id, "duration", time.Since(now))
	return mp, nil
}

func (mm *MultiManager) DeleteMultiProxy(id uuid.UUID) error {
	return mm.deleteMultiProxy(id, true)
}

func (mm *MultiManager) deleteMultiProxy(id uuid.UUID, first bool) error {
	now := time.Now()

	mm.mu.Lock()
	delete(mm.proxyMap, id)
	mm.mu.Unlock()

	if first {
		err := mm.db.DeleteProxyData(id)
		if err != nil {
			return err
		}

		m := mm.ownerMP.GetId().String() + "_" + id.String() + "_delete"
		err = mm.db.Publish(multi.UpdateMultiProxyChannel, m)
		if err != nil {
			return err
		}
	}

	mm.l.Info("deleted multiproxy", "proxyId", id, "duration", time.Since(now))
	return nil
}

func (mm *MultiManager) GetMultiProxy(id uuid.UUID) (*multi.Proxy, error) {
	mm.mu.RLock()
	mp, ok := mm.proxyMap[id]
	mm.mu.RUnlock()

	if ok {
		return mp, nil
	}

	return mm.CreateMultiProxyFromDatabase(id)
}

func (mm *MultiManager) CreateMultiProxyFromDatabase(id uuid.UUID) (*multi.Proxy, error) {
	data, err := mm.db.GetProxyData(id)
	if err != nil {
		return nil, err
	}

	var managerId uuid.UUID
	if mm.ownerMP == nil {
		managerId = id
	} else {
		managerId = mm.ownerMP.GetId()
	}

	mp := multi.NewProxy(id, managerId, mm.l, mm.db, mm.cf, data)

	mm.mu.Lock()
	mm.proxyMap[id] = mp
	mm.mu.Unlock()

	return mp, nil
}

func (mm *MultiManager) GetAllMultiProxies() []*multi.Proxy {
	var l []*multi.Proxy

	mm.mu.RLock()
	for _, mp := range mm.proxyMap {
		l = append(l, mp)
	}
	mm.mu.RUnlock()

	return l
}

func (mm *MultiManager) GetAllMultiProxiesFromDatabase() ([]*multi.Proxy, error) {
	var l []*multi.Proxy

	i, err := mm.db.GetAllProxyIds()
	if err != nil {
		return nil, err
	}

	for _, id := range i {
		mp, err := mm.GetMultiProxy(id)
		if err != nil {
			return nil, err
		}

		l = append(l, mp)
	}

	return l, nil
}

// can return nil if no other proxy is found
func (mm *MultiManager) GetProxyWithLowestPlayerCount(includingThisProxy bool) *multi.Proxy {
	var count int = math.MaxUint32
	var proxy *multi.Proxy

	for _, p := range mm.proxyMap {
		if !includingThisProxy {
			if p == mm.ownerMP {
				continue
			}
		}

		c := len(p.GetPlayerIds())
		if c < count {
			proxy = p
			count = c
		}
	}

	return proxy
}
