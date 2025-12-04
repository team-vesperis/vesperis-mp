package key

import (
	"errors"
	"slices"
)

type PlayerKey string

func (pk PlayerKey) String() string {
	return string(pk)
}

var ErrIncorrectPlayerKey = errors.New("incorrect player key")

const (
	PlayerKey_Proxy    PlayerKey = "proxy"
	PlayerKey_Backend  PlayerKey = "backend"
	PlayerKey_Username PlayerKey = "username"
	PlayerKey_Nickname PlayerKey = "nickname"

	PlayerKey_Friend_Friends               PlayerKey = "friend.friends"
	PlayerKey_Friend_FriendRequests        PlayerKey = "friend.friendRequests"
	PlayerKey_Friend_FriendPendingRequests PlayerKey = "friend.friendPendingRequests"

	PlayerKey_Party_IsInParty     PlayerKey = "party.isInParty"
	PlayerKey_Party_PartyOwner    PlayerKey = "party.partyOwner"
	PlayerKey_Party_Party         PlayerKey = "party.party"
	PlayerKey_Party_PartyRequests PlayerKey = "party.partyRequests"
	PlayerKey_Party_PartyInvites  PlayerKey = "party.partyInvites"

	PlayerKey_Permission_Role PlayerKey = "permission.role"
	PlayerKey_Permission_Rank PlayerKey = "permission.rank"

	PlayerKey_Ban_Banned      PlayerKey = "ban.banned"
	PlayerKey_Ban_Reason      PlayerKey = "ban.reason"
	PlayerKey_Ban_Permanently PlayerKey = "ban.permanently"
	PlayerKey_Ban_Expiration  PlayerKey = "ban.expiration"

	PlayerKey_Online   PlayerKey = "online"
	PlayerKey_Vanished PlayerKey = "vanished"
	PlayerKey_LastSeen PlayerKey = "lastSeen"
)

var AllowedPlayerKeys = []PlayerKey{
	PlayerKey_Proxy,
	PlayerKey_Backend,
	PlayerKey_Username,
	PlayerKey_Nickname,

	PlayerKey_Friend_Friends,
	PlayerKey_Friend_FriendRequests,
	PlayerKey_Friend_FriendPendingRequests,

	PlayerKey_Party_IsInParty,
	PlayerKey_Party_PartyOwner,
	PlayerKey_Party_Party,
	PlayerKey_Party_PartyRequests,
	PlayerKey_Party_PartyInvites,

	PlayerKey_Permission_Role,
	PlayerKey_Permission_Rank,

	PlayerKey_Ban_Banned,
	PlayerKey_Ban_Reason,
	PlayerKey_Ban_Permanently,
	PlayerKey_Ban_Expiration,

	PlayerKey_Online,
	PlayerKey_Vanished,
	PlayerKey_LastSeen,
}

func GetPlayerKey(s string) (PlayerKey, error) {
	pk := PlayerKey(s)
	if !slices.Contains(AllowedPlayerKeys, pk) {
		return PlayerKey(""), ErrIncorrectPlayerKey
	}

	return pk, nil
}
