package manager

import (
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (mm *MultiManager) StartBackend() {
	// start update listener
	mm.db.CreateListener(multi.UpdateMultiBackendChannel, mm.createBackendUpdateListener())

	// fill map
	_, err := mm.GetAllMultiBackendsFromDatabase()
	if err != nil {
		mm.l.Error("filling up multibackend map error", "error", err)
	}
}

func (mm *MultiManager) createBackendUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		m := msg.Payload
		mm.l.Debug("received backend update request", "message", m)

		s := strings.Split(m, "_")

		originProxy := s[0]
		// from own proxy, no update needed
		if mm.ownerMP.GetId().String() == originProxy {
			return
		}

		id, err := uuid.Parse(s[1])
		if err != nil {
			mm.l.Error("multibackend update channel parse uuid error", "parsed uuid", s[1], "error", err)
			return
		}

		k := s[2]

		mb, err := mm.GetMultiBackend(id)
		if err != nil {
			mm.l.Error("multibackend update channel get multibackend error", "backendId", id, "error", err)
			return
		}

		// already created
		if k == "new" {
			return
		}

		if k == "delete" {
			err := mm.DeleteMultiBackend(id)
			if err != nil {
				mm.l.Error("multibackend update channel delete multibackend error", "backendId", id, "error", err)
			}
			return
		}

		dataKey, err := key.GetBackendKey(k)
		if err != nil {
			mm.l.Error("multibackend update channel get data key error", "backendId", id, "key", k, "error", err)
			return
		}

		mb.Update(dataKey)
	}
}

func (mm *MultiManager) NewMultiBackend(addr string, id uuid.UUID) (*multi.Backend, error) {
	now := time.Now()
	mm.l.Info("creating new backend")

	data := &data.BackendData{
		Proxy:       mm.ownerMP.GetId(),
		Address:     addr,
		Maintenance: false,
		Players:     make([]uuid.UUID, 0),
	}

	err := mm.db.SetBackendData(id, data)
	if err != nil {
		return nil, err
	}

	mb, err := mm.CreateMultiBackendFromDatabase(id)
	if err != nil {
		return nil, err
	}

	err = mm.ownerMP.AddBackend(mb.GetId())
	if err != nil {
		return nil, err
	}

	mm.l.Info("new list", "list", mm.ownerMP.GetBackendsIds())

	m := mm.ownerMP.GetId().String() + "_" + id.String() + "_new"
	err = mm.db.Publish(multi.UpdateMultiProxyChannel, m)
	if err != nil {
		return nil, err
	}

	mm.l.Info("created new multibackend", "backendId", id, "duration", time.Since(now))
	return mb, nil
}

func (mm *MultiManager) DeleteMultiBackend(id uuid.UUID) error {
	now := time.Now()
	for key := range mm.backendMap {
		if key == id {
			mm.backendMap[key] = nil
		}
	}

	_, err := mm.db.GetBackendData(id)
	if err != nil {
		if err == database.ErrDataNotFound {
			return nil
		}

		mm.l.Error("could not get backend data")
		return err
	}

	err = mm.db.DeleteBackendData(id)
	if err != nil {
		return err
	}

	err = mm.ownerMP.RemoveBackendId(id)
	if err != nil {
		return err
	}

	m := mm.ownerMP.GetId().String() + "_" + id.String() + "_delete"
	err = mm.db.Publish(multi.UpdateMultiBackendChannel, m)
	if err != nil {
		return err
	}

	mm.l.Info("deleted multibackend", "backendId", id, "duration", time.Since(now))
	return nil
}

func (mm *MultiManager) GetMultiBackend(id uuid.UUID) (*multi.Backend, error) {
	mm.mu.RLock()
	mb, ok := mm.backendMap[id]
	mm.mu.RUnlock()

	if ok {
		return mb, nil
	}

	return mm.CreateMultiBackendFromDatabase(id)
}

func (mm *MultiManager) GetMultiBackendUsingAddress(addr string) (*multi.Backend, error) {
	l := mm.GetAllMultiBackends()
	for _, mb := range l {
		if mb.GetAddress() == addr {
			return mb, nil
		}
	}

	l, err := mm.GetAllMultiBackendsFromDatabase()
	if err != nil {
		return nil, err
	}

	for _, mb := range l {
		if mb.GetAddress() == addr {
			return mb, nil
		}
	}

	return nil, multi.ErrBackendNotFound
}

func (mm *MultiManager) CreateMultiBackendFromDatabase(id uuid.UUID) (*multi.Backend, error) {
	data, err := mm.db.GetBackendData(id)
	if err != nil {
		return nil, err
	}

	mp, err := mm.GetMultiProxy(data.Proxy)
	if err != nil {
		return nil, err
	}

	mb := multi.NewBackend(id, mm.ownerMP.GetId(), mp, mm.l, mm.db, mm.cf, data)

	mm.mu.Lock()
	mm.backendMap[id] = mb
	mm.mu.Unlock()

	return mb, nil
}

func (mm *MultiManager) GetAllMultiBackends() []*multi.Backend {
	var l []*multi.Backend

	mm.mu.RLock()
	for _, mb := range mm.backendMap {
		l = append(l, mb)
	}
	mm.mu.RUnlock()

	return l
}

func (mm *MultiManager) GetAllMultiBackendsFromDatabase() ([]*multi.Backend, error) {
	var l []*multi.Backend

	i, err := mm.db.GetAllBackendsIds()
	if err != nil {
		return nil, err
	}

	for _, id := range i {
		mb, err := mm.GetMultiBackend(id)
		if err != nil {
			return nil, err
		}

		l = append(l, mb)
	}

	return l, nil
}

// creates id
func (mm *MultiManager) CreateNewBackendId() (uuid.UUID, error) {
	var break_err error

	for {
		id := uuid.New()
		_, err := mm.GetMultiBackend(id)
		if err == database.ErrDataNotFound {
			return id, nil
		}

		if err != nil {
			break_err = err
			break
		}
	}

	mm.l.Error("create new backend id error", "error", break_err)
	return uuid.Nil, break_err
}
