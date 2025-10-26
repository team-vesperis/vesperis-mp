package key

import (
	"errors"
	"slices"
)

type BackendKey string

func (pk BackendKey) String() string {
	return string(pk)
}

var ErrIncorrectBackendKey = errors.New("incorrect backend key")

const (
	BackendKey_Maintenance BackendKey = "maintenance"
	BackendKey_PlayerList  BackendKey = "players"
)

var AllowedBackendKeys = []BackendKey{
	BackendKey_Maintenance,
	BackendKey_PlayerList,
}

func GetBackendKey(s string) (BackendKey, error) {
	pk := BackendKey(s)
	if !slices.Contains(AllowedBackendKeys, pk) {
		return BackendKey(""), ErrIncorrectBackendKey
	}

	return pk, nil
}
