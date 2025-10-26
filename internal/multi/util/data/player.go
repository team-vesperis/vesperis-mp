package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"go.minekube.com/gate/pkg/util/uuid"
)

type PlayerData struct {
	ProxyId    uuid.UUID       `json:"proxy_id,omitempty"`
	BackendId  uuid.UUID       `json:"backend_id,omitempty"`
	Username   string          `json:"username,omitempty"`
	Nickname   string          `json:"nickname,omitempty"`
	Permission *PermissionData `json:"permission,omitempty"`
	Ban        *BanData        `json:"ban,omitempty"`
	Online     bool            `json:"online,omitempty"`
	Vanished   bool            `json:"vanished,omitempty"`
	LastSeen   *time.Time      `json:"last_seen,omitempty"`
	Friends    []uuid.UUID     `json:"friends,omitempty"`
}

type PermissionData struct {
	Role string `json:"role,omitempty"`
	Rank string `json:"rank,omitempty"`
}

type BanData struct {
	Banned      bool      `json:"banned,omitempty"`
	Reason      string    `json:"reason,omitempty"`
	Permanently bool      `json:"permanently,omitempty"`
	Expiration  time.Time `json:"expiration,omitempty"`
}

func (pd PlayerData) Value() (driver.Value, error) {
	return json.Marshal(pd)
}

func (pd *PlayerData) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, pd)
	case string:
		return json.Unmarshal([]byte(v), pd)
	case nil:
		*pd = PlayerData{}
		return nil
	default:
		return errors.New("unsupported type for PlayerData Scan")
	}
}
