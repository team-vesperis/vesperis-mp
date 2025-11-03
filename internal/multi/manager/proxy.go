package manager

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (mm *MultiManager) StartProxy() {
	// start update listener
	mm.db.CreateListener(multi.UpdateMultiProxyChannel, mm.createProxyUpdateListener())

	// fill map
	_, err := mm.GetAllMultiProxiesFromDatabase()
	if err != nil {
		mm.l.Error("filling up multiproxy map error", "error", err)
	}
}

func (mm *MultiManager) createProxyUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		mm.l.Info(msg.Payload)
		m := msg.Payload
		s := strings.Split(m, "_")

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

		// already created
		if k == "new" {
			return
		}

		if k == "delete" {
			err := mm.DeleteMultiProxy(id)
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

func (mm *MultiManager) NewMultiProxy(id uuid.UUID) (*multi.Proxy, error) {
	now := time.Now()

	addr := fmt.Sprintf("%s.proxy.default.svc.cluster.local:25565", id.String())

	data := &data.ProxyData{
		Address:     addr,
		Maintenance: false,
		Backends:    make([]uuid.UUID, 0),
		Players:     make([]uuid.UUID, 0),
	}

	err := mm.db.SetProxyData(id, data)
	if err != nil {
		return nil, err
	}

	mp, err := mm.CreateMultiProxyFromDatabase(id)
	if err != nil {
		return nil, err
	}

	// update every proxies' map
	if mm.ownerMP == nil {
		mm.ownerMP = mp
	}

	m := mm.ownerMP.GetId().String() + "_" + id.String() + "_new"
	err = mm.db.Publish(multi.UpdateMultiProxyChannel, m)
	if err != nil {
		return nil, err
	}

	mm.l.Info("created new multiproxy", "proxyId", id, "duration", time.Since(now))
	return mp, nil
}

func (mm *MultiManager) DeleteMultiProxy(id uuid.UUID) error {
	now := time.Now()
	for key := range mm.proxyMap {
		if key == id {
			mm.proxyMap[key] = nil
		}
	}

	_, err := mm.db.GetProxyData(id)
	if err != nil {
		if err == database.ErrDataNotFound {
			return nil
		}

		mm.l.Error("could not get proxy data")
		return err
	}

	err = mm.db.DeleteProxyData(id)
	if err != nil {
		return err
	}

	m := mm.ownerMP.GetId().String() + "_" + id.String() + "_delete"
	err = mm.db.Publish(multi.UpdateMultiProxyChannel, m)
	if err != nil {
		return err
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

	mp := multi.NewProxy(id, managerId, mm.db, data)

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

// creates id
func (mm *MultiManager) CreateNewProxyId() (uuid.UUID, error) {
	var break_err error

	for {
		id := uuid.New()
		_, err := mm.GetMultiProxy(id)
		if err == database.ErrDataNotFound {
			return id, nil
		}

		if err != nil {
			break_err = err
			break
		}
	}

	mm.l.Error("create new proxy id error", "error", break_err)
	return uuid.Nil, break_err
}

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
