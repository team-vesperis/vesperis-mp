package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"go.minekube.com/gate/pkg/util/uuid"
)

type ProxyData struct {
	Address       string      `json:"address"`
	Maintenance   bool        `json:"maintenance"`
	Backends      []uuid.UUID `json:"backends"`
	Players       []uuid.UUID `json:"players"`
	LastHeartBeat *time.Time  `json:"lastHartBeat"`
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
