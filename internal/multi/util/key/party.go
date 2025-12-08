package key

import (
	"errors"
	"slices"
)

type PartyKey string

func (pk PartyKey) String() string {
	return string(pk)
}

var ErrIncorrectPartyKey = errors.New("incorrect party key")

const (
	PartyKey_PartyOwner        PartyKey = "partyOwner"
	PartyKey_PartyMembers      PartyKey = "partyMembers"
	PartyKey_PartyJoinRequests PartyKey = "partyJoinRequests"
	PartyKey_PartyInvitations  PartyKey = "partyInvitations"
)

var AllowedPartyKeys = []PartyKey{
	PartyKey_PartyOwner,
	PartyKey_PartyMembers,
	PartyKey_PartyJoinRequests,
	PartyKey_PartyInvitations,
}

func GetPartyKey(s string) (PartyKey, error) {
	pk := PartyKey(s)
	if !slices.Contains(AllowedPartyKeys, pk) {
		return PartyKey(""), ErrIncorrectBackendKey
	}

	return pk, nil
}
