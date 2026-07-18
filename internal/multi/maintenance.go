package multi

import (
	"sync"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
)

type maintenanceInfo struct {
	inMaintenance bool
	reason        string
	expire        bool
	expiration    time.Time

	mu    sync.RWMutex
	owner Owner
}

// either a proxy or backend
type Owner interface {
	save(key.Key, any) error
}

func newMaintenanceInfo(owner Owner, m data.MaintenanceData) *maintenanceInfo {
	return &maintenanceInfo{
		inMaintenance: m.InMaintenance,
		reason:        m.Reason,
		expire:        m.Expire,
		expiration:    m.Expiration,

		mu:    sync.RWMutex{},
		owner: owner,
	}
}

func (mi *maintenanceInfo) IsInMaintenance() bool {
	mi.mu.RLock()
	defer mi.mu.RUnlock()

	return mi.inMaintenance
}

func (mi *maintenanceInfo) SetMaintenance(maintenance bool) error {
	return mi.setMaintenance(maintenance, true)
}

func (mi *maintenanceInfo) setMaintenance(maintenance, notify bool) error {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	mi.inMaintenance = maintenance

	if notify {
		return mi.owner.save(key.MaintenanceKey_InMaintenance, maintenance)
	}

	return nil
}

func (mi *maintenanceInfo) GetReason() string {
	mi.mu.RLock()
	defer mi.mu.RUnlock()

	return mi.reason
}

func (mi *maintenanceInfo) SetReason(reason string) error {
	return mi.setReason(reason, true)
}

func (mi *maintenanceInfo) setReason(reason string, notify bool) error {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	mi.reason = reason

	if notify {
		return mi.owner.save(key.MaintenanceKey_Reason, reason)
	}

	return nil
}

func (mi *maintenanceInfo) WillExpire() bool {
	mi.mu.RLock()
	defer mi.mu.RUnlock()

	return mi.expire
}

func (mi *maintenanceInfo) SetExpire(expire bool) error {
	return mi.setExpire(expire, true)
}

func (mi *maintenanceInfo) setExpire(expire, notify bool) error {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	mi.expire = expire

	if notify {
		return mi.owner.save(key.MaintenanceKey_Expire, expire)
	}

	return nil
}

func (mi *maintenanceInfo) GetExpiration() time.Time {
	mi.mu.RLock()
	defer mi.mu.RUnlock()

	return mi.expiration
}

func (mi *maintenanceInfo) SetExpiration(expiration time.Time) error {
	return mi.setExpiration(expiration, true)
}

func (mi *maintenanceInfo) setExpiration(expiration time.Time, notify bool) error {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	mi.expiration = expiration

	if notify {
		return mi.owner.save(key.MaintenanceKey_Expiration, expiration)
	}

	return nil
}
