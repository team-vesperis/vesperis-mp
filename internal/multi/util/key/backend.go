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
	BackendKey_PlayerList                BackendKey = "players"
	BackendKey_Maintenance_InMaintenance BackendKey = "maintenance.inMaintenance"
	BackendKey_Maintenance_Reason        BackendKey = "maintenance.reason"
	BackendKey_Maintenance_Expire        BackendKey = "maintenance.expire"
	BackendKey_Maintenance_Expiration    BackendKey = "maintenance.expiration"
)

var AllowedBackendKeys = []BackendKey{
	BackendKey_PlayerList,
	BackendKey_Maintenance_InMaintenance,
	BackendKey_Maintenance_Reason,
	BackendKey_Maintenance_Expire,
	BackendKey_Maintenance_Expiration,
}

func GetBackendKey(s string) (BackendKey, error) {
	bk := BackendKey(s)
	if !slices.Contains(AllowedBackendKeys, bk) {
		return BackendKey(""), ErrIncorrectBackendKey
	}

	return bk, nil
}
