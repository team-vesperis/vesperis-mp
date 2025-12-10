package multi

import (
	"errors"
	"slices"
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

type Party struct {
	id uuid.UUID

	partyOwner uuid.UUID

	partyMembers      []uuid.UUID
	partyJoinRequests []uuid.UUID
	partyInvitations  []uuid.UUID

	managerId uuid.UUID
	l         *logger.Logger
	db        *database.Database
	mu        sync.RWMutex
}

func NewParty(id, managerId uuid.UUID, l *logger.Logger, db *database.Database, data *data.PartyData) *Party {
	mp := &Party{
		id:        id,
		managerId: managerId,
		l:         l,
		db:        db,
		mu:        sync.RWMutex{},
	}

	mp.partyOwner = data.PartyOwner
	mp.partyMembers = data.PartyMembers
	mp.partyJoinRequests = data.PartyJoinRequests
	mp.partyInvitations = data.PartyInvitations

	return mp
}

var ErrPartyNotFound = errors.New("party not found")

const UpdateMultiPartyChannel = "update_multiparty"

func (mp *Party) save(k key.PartyKey, val any) error {
	err := mp.db.SetPartyDataField(mp.id, k, val)
	if err != nil {
		return err
	}

	m := mp.managerId.String() + "_" + mp.id.String() + "_" + k.String()
	return mp.db.Publish(UpdateMultiPartyChannel, m)
}

func (mp *Party) Update(k key.PartyKey) {
	var err error

	switch k {
	case key.PartyKey_PartyOwner:
		var owner uuid.UUID
		err = mp.db.GetPartyDataField(mp.id, key.PartyKey_PartyOwner, &owner)
		mp.setPartyOwner(owner, false)
	case key.PartyKey_PartyMembers:
		var members []uuid.UUID
		err = mp.db.GetPartyDataField(mp.id, key.PartyKey_PartyMembers, &members)
		mp.setPartyMembers(members, false)
	case key.PartyKey_PartyJoinRequests:
		var requests []uuid.UUID
		err = mp.db.GetPartyDataField(mp.id, key.PartyKey_PartyJoinRequests, &requests)
		mp.setPartyJoinRequests(requests, false)
	case key.PartyKey_PartyInvitations:
		var invitations []uuid.UUID
		err = mp.db.GetPartyDataField(mp.id, key.PartyKey_PartyInvitations, &invitations)
		mp.setPartyInvitations(invitations, false)
	}

	if err != nil {
		mp.l.Error("multiparty update partykey get field from database error", "error", err)
	}
}

func (mp *Party) GetId() uuid.UUID {
	return mp.id
}

func (mp *Party) GetPartyOwner() uuid.UUID {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.partyOwner
}

func (mp *Party) SetPartyOwner(owner uuid.UUID) error {
	return mp.setPartyOwner(owner, true)
}

func (mp *Party) setPartyOwner(owner uuid.UUID, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.partyOwner = owner

	if notify {
		return mp.save(key.PartyKey_PartyOwner, owner)
	}

	return nil
}

func (mp *Party) GetPartyMembers() []uuid.UUID {
	mp.mu.RLock()
	c := append([]uuid.UUID{}, mp.partyMembers...)
	mp.mu.RUnlock()

	return c
}

func (mp *Party) SetPartyMembers(ids []uuid.UUID) error {
	return mp.setPartyMembers(ids, true)
}

func (mp *Party) setPartyMembers(ids []uuid.UUID, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.partyMembers = ids

	if notify {
		return mp.save(key.PartyKey_PartyMembers, ids)
	}

	return nil
}

func (mp *Party) AddPartyMember(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !slices.Contains(mp.partyMembers, id) {
		mp.partyMembers = append(mp.partyMembers, id)
	}

	return mp.save(key.PartyKey_PartyMembers, mp.partyMembers)
}

func (mp *Party) RemovePartyMember(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	i := slices.Index(mp.partyMembers, id)
	if i == -1 {
		return ErrPlayerNotFound
	}
	mp.partyMembers = slices.Delete(mp.partyMembers, i, i+1)

	return mp.save(key.PartyKey_PartyMembers, mp.partyMembers)
}

func (mp *Party) GetPartyJoinRequests() []uuid.UUID {
	mp.mu.RLock()
	c := append([]uuid.UUID{}, mp.partyJoinRequests...)
	mp.mu.RUnlock()

	return c
}

func (mp *Party) SetPartyJoinRequests(ids []uuid.UUID) error {
	return mp.setPartyJoinRequests(ids, true)
}

func (mp *Party) setPartyJoinRequests(ids []uuid.UUID, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.partyJoinRequests = ids

	if notify {
		return mp.save(key.PartyKey_PartyJoinRequests, ids)
	}

	return nil
}

func (mp *Party) AddPartyJoinRequest(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !slices.Contains(mp.partyJoinRequests, id) {
		mp.partyJoinRequests = append(mp.partyJoinRequests, id)
	}

	return mp.save(key.PartyKey_PartyJoinRequests, mp.partyJoinRequests)
}

func (mp *Party) RemovePartyJoinRequest(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	i := slices.Index(mp.partyJoinRequests, id)
	if i == -1 {
		return ErrPlayerNotFound
	}
	mp.partyJoinRequests = slices.Delete(mp.partyJoinRequests, i, i+1)

	return mp.save(key.PartyKey_PartyJoinRequests, mp.partyJoinRequests)
}

func (mp *Party) GetPartyInvitations() []uuid.UUID {
	mp.mu.RLock()
	c := append([]uuid.UUID{}, mp.partyInvitations...)
	mp.mu.RUnlock()

	return c
}

func (mp *Party) SetPartyInvitations(ids []uuid.UUID) error {
	return mp.setPartyInvitations(ids, true)
}

func (mp *Party) setPartyInvitations(ids []uuid.UUID, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.partyInvitations = ids

	if notify {
		return mp.save(key.PartyKey_PartyInvitations, ids)
	}

	return nil
}

func (mp *Party) AddPartyInvitation(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !slices.Contains(mp.partyInvitations, id) {
		mp.partyInvitations = append(mp.partyInvitations, id)
	}

	return mp.save(key.PartyKey_PartyInvitations, mp.partyInvitations)
}

func (mp *Party) RemovePartyInvitation(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	i := slices.Index(mp.partyInvitations, id)
	if i == -1 {
		return ErrPlayerNotFound
	}
	mp.partyInvitations = slices.Delete(mp.partyInvitations, i, i+1)

	return mp.save(key.PartyKey_PartyInvitations, mp.partyInvitations)
}
