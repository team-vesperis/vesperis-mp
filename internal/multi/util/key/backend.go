package key

import (
	"errors"
	"slices"
)

type BackendKey string

func (bk BackendKey) String() string {
	return string(bk)
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
	bk := BackendKey(s)
	if !slices.Contains(AllowedBackendKeys, bk) {
		return BackendKey(""), ErrIncorrectBackendKey
	}

	return bk, nil
}
