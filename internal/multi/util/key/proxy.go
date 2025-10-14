package key

import (
	"errors"
	"slices"
)

type ProxyKey string

func (pk ProxyKey) String() string {
	return string(pk)
}

var ErrIncorrectProxyKey = errors.New("incorrect proxy key")

const (
	ProxyKey_Maintenance ProxyKey = "maintenance"
	ProxyKey_BackendList ProxyKey = "backends"
	ProxyKey_PlayerList  ProxyKey = "players"
)

var AllowedProxyKeys = []ProxyKey{
	ProxyKey_Maintenance,
	ProxyKey_BackendList,
	ProxyKey_PlayerList,
}

func GetProxyKey(s string) (ProxyKey, error) {
	pk := ProxyKey(s)
	if !slices.Contains(AllowedProxyKeys, pk) {
		return ProxyKey(""), ErrIncorrectProxyKey
	}

	return pk, nil
}
