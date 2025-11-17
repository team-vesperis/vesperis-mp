package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"go.minekube.com/gate/pkg/util/uuid"
)

type BackendData struct {
	Address     string      `json:"address"`
	Proxy       uuid.UUID   `json:"proxy"`
	Maintenance bool        `json:"maintenance"`
	Players     []uuid.UUID `json:"players"`
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
