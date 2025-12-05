package multi

import (
	"slices"
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

type partyInfo struct {
	isInParty  bool
	partyOwner uuid.UUID

	party             []uuid.UUID
	partyJoinRequests []uuid.UUID
	partyInvites      []uuid.UUID

	mu sync.RWMutex
	mp *Player
}

func newPartyInfo(mp *Player, data *data.PlayerData) *partyInfo {
	return &partyInfo{
		isInParty:         data.Party.IsInParty,
		partyOwner:        data.Party.PartyOwner,
		party:             data.Party.Party,
		partyJoinRequests: data.Party.PartyJoinRequests,
		partyInvites:      data.Party.PartyInvites,

		mu: sync.RWMutex{},
		mp: mp,
	}
}

func (pi *partyInfo) IsInParty() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.isInParty
}

func (pi *partyInfo) SetIsInParty(inParty bool) error {
	return pi.setIsInParty(inParty, true)
}

func (pi *partyInfo) setIsInParty(inParty, notify bool) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.isInParty = inParty

	if notify {
		return pi.mp.save(key.PlayerKey_Party_IsInParty, inParty)
	}

	return nil
}

func (pi *partyInfo) GetPartyOwner() uuid.UUID {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.partyOwner
}

func (pi *partyInfo) SetPartyOwner(owner uuid.UUID) error {
	return pi.setPartyOwner(owner, true)
}

func (pi *partyInfo) setPartyOwner(owner uuid.UUID, notify bool) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.partyOwner = owner

	if notify {
		return pi.mp.save(key.PlayerKey_Party_PartyOwner, owner)
	}

	return nil
}

func (pi *partyInfo) GetPartyIds() []uuid.UUID {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return slices.Clone(pi.party)
}

func (pi *partyInfo) SetPartyIds(ids []uuid.UUID) error {
	return pi.setPartyIds(ids, true)
}

func (pi *partyInfo) setPartyIds(ids []uuid.UUID, notify bool) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.party = ids

	if notify {
		return pi.mp.save(key.PlayerKey_Party_Party, ids)
	}

	return nil
}

func (pi *partyInfo) IsInPartyMember(id uuid.UUID) bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return slices.Contains(pi.party, id)
}

func (pi *partyInfo) AddPartyMemberId(id uuid.UUID) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if !slices.Contains(pi.party, id) {
		pi.party = append(pi.party, id)
	}

	return pi.mp.save(key.PlayerKey_Party_Party, pi.party)
}

func (pi *partyInfo) RemovePartyMemberId(id uuid.UUID) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	i := slices.Index(pi.party, id)
	if i == -1 {
		return ErrPlayerNotFound
	}
	pi.party = slices.Delete(pi.party, i, i+1)

	return pi.mp.save(key.PlayerKey_Party_Party, pi.party)
}

func (pi *partyInfo) GetPartyJoinRequestIds() []uuid.UUID {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return slices.Clone(pi.partyJoinRequests)
}

func (pi *partyInfo) SetPartyJoinRequestIds(ids []uuid.UUID) error {
	return pi.setPartyJoinRequestIds(ids, true)
}

func (pi *partyInfo) setPartyJoinRequestIds(ids []uuid.UUID, notify bool) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.partyJoinRequests = ids

	if notify {
		return pi.mp.save(key.PlayerKey_Party_PartyJoinRequests, ids)
	}

	return nil
}

func (pi *partyInfo) IsPartyJoinRequest(id uuid.UUID) bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return slices.Contains(pi.partyJoinRequests, id)
}

func (pi *partyInfo) AddPartyJoinRequestId(id uuid.UUID) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if !slices.Contains(pi.partyJoinRequests, id) {
		pi.partyJoinRequests = append(pi.partyJoinRequests, id)
	}

	return pi.mp.save(key.PlayerKey_Party_PartyJoinRequests, pi.partyJoinRequests)
}

func (pi *partyInfo) RemovePartyJoinRequestId(id uuid.UUID) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	i := slices.Index(pi.partyJoinRequests, id)
	if i == -1 {
		return ErrPlayerNotFound
	}
	pi.partyJoinRequests = slices.Delete(pi.partyJoinRequests, i, i+1)

	return pi.mp.save(key.PlayerKey_Party_PartyJoinRequests, pi.partyJoinRequests)
}

func (pi *partyInfo) GetPartyInviteIds() []uuid.UUID {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return slices.Clone(pi.partyInvites)
}

func (pi *partyInfo) SetPartyInviteIds(ids []uuid.UUID) error {
	return pi.setPartyInviteIds(ids, true)
}

func (pi *partyInfo) setPartyInviteIds(ids []uuid.UUID, notify bool) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.partyInvites = ids

	if notify {
		return pi.mp.save(key.PlayerKey_Party_PartyInvites, ids)
	}

	return nil
}

func (pi *partyInfo) IsPartyInvite(id uuid.UUID) bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return slices.Contains(pi.partyInvites, id)
}

func (pi *partyInfo) AddPartyInviteId(id uuid.UUID) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	if !slices.Contains(pi.partyInvites, id) {
		pi.partyInvites = append(pi.partyInvites, id)
	}

	return pi.mp.save(key.PlayerKey_Party_PartyInvites, pi.partyInvites)
}

func (pi *partyInfo) RemovePartyInviteId(id uuid.UUID) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	i := slices.Index(pi.partyInvites, id)
	if i == -1 {
		return ErrPlayerNotFound
	}
	pi.partyInvites = slices.Delete(pi.partyInvites, i, i+1)

	return pi.mp.save(key.PlayerKey_Party_PartyInvites, pi.partyInvites)
}
