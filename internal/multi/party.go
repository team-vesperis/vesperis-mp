package multi

import (
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"go.minekube.com/gate/pkg/util/uuid"
)

type partyInfo struct {
	isInParty  bool
	partyOwner uuid.UUID

	party         []uuid.UUID
	partyRequests []uuid.UUID
	partyInvites  []uuid.UUID

	mu sync.RWMutex
	mp *Player
}

func newPartyInfo(mp *Player, data *data.PlayerData) *partyInfo {
	return &partyInfo{
		isInParty:     data.Party.IsInParty,
		partyOwner:    data.Party.PartyOwner,
		party:         data.Party.Party,
		partyRequests: data.Party.PartyRequests,
		partyInvites:  data.Party.PartyInvites,

		mu: sync.RWMutex{},
		mp: mp,
	}
}
