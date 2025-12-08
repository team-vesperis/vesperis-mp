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

func (mm *MultiManager) createPartyUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		m := msg.Payload
		s := strings.Split(m, "_")
		if len(s) != 3 {
			mm.l.Warn("multiparty update channel received message with incorrect length", "message", m)
			return
		}

		originProxy := s[0]
		// from own proxy, no update needed
		if mm.ownerMP.GetId().String() == originProxy {
			return
		}

		id, err := uuid.Parse(s[1])
		if err != nil {
			mm.l.Error("multiparty update channel parse uuid error", "parsed uuid", s[1], "error", err)
			return
		}

		k := s[2]

		mp, err := mm.GetMultiParty(id)
		if err != nil {
			mm.l.Error("multiparty update channel get multiparty error", "partyId", id, "error", err)
			return
		}

		mm.l.Debug("received backend update request", "originProxyId", originProxy, "partyId", id, "key", k)

		// already created
		if k == "new" {
			return
		}

		if k == "delete" {
			err := mm.deleteMultiParty(id, false)
			if err != nil {
				mm.l.Error("multiparty update channel delete multiparty error", "partyId", id, "error", err)
			}
			return
		}

		dataKey, err := key.GetPartyKey(k)
		if err != nil {
			mm.l.Error("multiparty update channel get data key error", "partyId", id, "key", k, "error", err)
			return
		}

		mp.Update(dataKey)
	}
}

func (mm *MultiManager) NewMultiParty(ownerPlayer uuid.UUID) (*multi.Party, error) {
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

	data := &data.PartyData{
		PartyOwner:        ownerPlayer,
		PartyMembers:      append(make([]uuid.UUID, 0), ownerPlayer),
		PartyInvitations:  make([]uuid.UUID, 0),
		PartyJoinRequests: make([]uuid.UUID, 0),
	}

	err = mm.db.SetPartyData(id, data)
	if err != nil {
		return nil, err
	}

	mp, err := mm.CreateMultiPartyFromDatabase(id)
	if err != nil {
		return nil, err
	}

	m := mm.ownerMP.GetId().String() + "_" + id.String() + "_new"
	err = mm.db.Publish(multi.UpdateMultiPartyChannel, m)
	if err != nil {
		return nil, err
	}

	mm.l.Info("created new multiparty", "partyId", id, "duration", time.Since(now))
	return mp, nil
}

func (mm *MultiManager) DeleteMultiParty(id uuid.UUID) error {
	return mm.deleteMultiParty(id, true)
}

func (mm *MultiManager) deleteMultiParty(id uuid.UUID, first bool) error {
	now := time.Now()

	_, err := mm.GetMultiParty(id)
	if err != nil {
		return err
	}

	mm.mu.Lock()
	delete(mm.partyMap, id)
	mm.mu.Unlock()

	if first {
		err = mm.db.DeletePartyData(id)
		if err != nil {
			return err
		}

		m := mm.ownerMP.GetId().String() + "_" + id.String() + "_delete"
		err = mm.db.Publish(multi.UpdateMultiPartyChannel, m)
		if err != nil {
			return err
		}
	}

	mm.l.Info("deleted multiparty", "partyId", id, "duration", time.Since(now))
	return nil
}

func (mm *MultiManager) GetMultiParty(id uuid.UUID) (*multi.Party, error) {
	mm.mu.RLock()
	mp, ok := mm.partyMap[id]
	mm.mu.RUnlock()

	if ok {
		return mp, nil
	}

	return mm.CreateMultiPartyFromDatabase(id)
}

func (mm *MultiManager) CreateMultiPartyFromDatabase(id uuid.UUID) (*multi.Party, error) {
	data, err := mm.db.GetPartyData(id)
	if err != nil {
		return nil, err
	}

	mp := multi.NewParty(id, mm.ownerMP.GetId(), mm.l, mm.db, data)

	mm.mu.Lock()
	mm.partyMap[id] = mp
	mm.mu.Unlock()

	return mp, nil
}
