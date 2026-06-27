package data

import "time"

// shared with proxy and backend
type MaintenanceData struct {
	InMaintenance bool      `json:"inMaintenance"`
	Reason        string    `json:"reason"`
	Expire        bool      `json:"expire"`
	Expiration    time.Time `json:"expiration"`
}
