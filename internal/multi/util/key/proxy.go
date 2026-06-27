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
	ProxyKey_BackendList               ProxyKey = "backends"
	ProxyKey_PlayerList                ProxyKey = "players"
	ProxyKey_LastHeartBeat             ProxyKey = "lastHartBeat"
	ProxyKey_Maintenance_InMaintenance ProxyKey = "maintenance.inMaintenance"
	ProxyKey_Maintenance_Reason        ProxyKey = "maintenance.reason"
	ProxyKey_Maintenance_Expire        ProxyKey = "maintenance.expire"
	ProxyKey_Maintenance_Expiration    ProxyKey = "maintenance.expiration"
)

var AllowedProxyKeys = []ProxyKey{
	ProxyKey_BackendList,
	ProxyKey_PlayerList,
	ProxyKey_LastHeartBeat,
	ProxyKey_Maintenance_InMaintenance,
	ProxyKey_Maintenance_Reason,
	ProxyKey_Maintenance_Expire,
	ProxyKey_Maintenance_Expiration,
}

func GetProxyKey(s string) (ProxyKey, error) {
	pk := ProxyKey(s)
	if !slices.Contains(AllowedProxyKeys, pk) {
		return ProxyKey(""), ErrIncorrectProxyKey
	}

	return pk, nil
}
