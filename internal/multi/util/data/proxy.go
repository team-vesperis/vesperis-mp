package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"go.minekube.com/gate/pkg/util/uuid"
)

type ProxyData struct {
	Address       string      `json:"address,omitempty"`
	Maintenance   bool        `json:"maintenance,omitempty"`
	Backends      []uuid.UUID `json:"backends,omitempty"`
	Players       []uuid.UUID `json:"players,omitempty"`
	LastHeartBeat *time.Time  `json:"last_hart_beat,omitempty"`
}

func (pd ProxyData) Value() (driver.Value, error) {
	return json.Marshal(pd)
}

func (pd *ProxyData) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, pd)
	case string:
		return json.Unmarshal([]byte(v), pd)
	case nil:
		*pd = ProxyData{}
		return nil
	default:
		return errors.New("unsupported type for ProxyData Scan")
	}
}
