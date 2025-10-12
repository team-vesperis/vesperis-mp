package util

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
	PlayerKey_ProxyId   PlayerKey = "proxy_id"
	PlayerKey_BackendId PlayerKey = "backend_id"
	PlayerKey_Username  PlayerKey = "username"
	PlayerKey_Nickname  PlayerKey = "nickname"

	PlayerKey_Permission_Role PlayerKey = "permission.role"
	PlayerKey_Permission_Rank PlayerKey = "permission.rank"

	PlayerKey_Ban_Banned      PlayerKey = "ban.banned"
	PlayerKey_Ban_Reason      PlayerKey = "ban.reason"
	PlayerKey_Ban_Permanently PlayerKey = "ban.permanently"
	PlayerKey_Ban_Expiration  PlayerKey = "ban.expiration"

	PlayerKey_Online   PlayerKey = "online"
	PlayerKey_Vanished PlayerKey = "vanished"
	PlayerKey_LastSeen PlayerKey = "last_seen"
	PlayerKey_Friends  PlayerKey = "friends"
)

var AllowedPlayerKeys = []PlayerKey{
	PlayerKey_ProxyId,
	PlayerKey_BackendId,
	PlayerKey_Username,
	PlayerKey_Nickname,

	PlayerKey_Permission_Role,
	PlayerKey_Permission_Rank,

	PlayerKey_Ban_Banned,
	PlayerKey_Ban_Reason,
	PlayerKey_Ban_Permanently,
	PlayerKey_Ban_Expiration,

	PlayerKey_Online,
	PlayerKey_Vanished,
	PlayerKey_LastSeen,
	PlayerKey_Friends,
}

func GetPlayerKey(s string) (PlayerKey, error) {
	pk := PlayerKey(s)
	if !slices.Contains(AllowedPlayerKeys, pk) {
		return PlayerKey(""), ErrIncorrectPlayerKey
	}

	return pk, nil
}
