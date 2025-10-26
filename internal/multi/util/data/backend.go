package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"go.minekube.com/gate/pkg/util/uuid"
)

type BackendData struct {
	Address     string      `json:"address,omitempty"`
	Proxy       uuid.UUID   `json:"proxy,omitempty"`
	Maintenance bool        `json:"maintenance,omitempty"`
	Players     []uuid.UUID `json:"players,omitempty"`
}

func (bd BackendData) Value() (driver.Value, error) {
	return json.Marshal(bd)
}

func (bd *BackendData) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, bd)
	case string:
		return json.Unmarshal([]byte(v), bd)
	case nil:
		*bd = BackendData{}
		return nil
	default:
		return errors.New("unsupported type for BackendData Scan")
	}
}
