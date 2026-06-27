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

func newProxyMaintenanceInfo(p *Proxy, data *data.ProxyData) *maintenanceInfo {
	return &maintenanceInfo{
		inMaintenance: data.Maintenance.InMaintenance,
		reason:        data.Maintenance.Reason,
		expire:        data.Maintenance.Expire,
		expiration:    data.Maintenance.Expiration,

		mu:    sync.RWMutex{},
		owner: p,
	}
}

func newBackendMaintenanceInfo(b *Backend, data *data.BackendData) *maintenanceInfo {
	return &maintenanceInfo{
		inMaintenance: data.Maintenance.InMaintenance,
		reason:        data.Maintenance.Reason,
		expire:        data.Maintenance.Expire,
		expiration:    data.Maintenance.Expiration,

		mu:    sync.RWMutex{},
		owner: b,
	}
}
