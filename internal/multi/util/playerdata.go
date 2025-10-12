package util

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"go.minekube.com/gate/pkg/util/uuid"
)

type PlayerData struct {
	ProxyId    uuid.UUID       `json:"proxy_id"`
	BackendId  uuid.UUID       `json:"backend_id"`
	Username   string          `json:"username"`
	Nickname   string          `json:"nickname"`
	Permission *PermissionData `json:"permission"`
	Ban        *BanData        `json:"ban"`
	Online     bool            `json:"online"`
	Vanished   bool            `json:"vanished"`
	LastSeen   *time.Time      `json:"last_seen"`
	Friends    []uuid.UUID     `json:"friends"`
}

type PermissionData struct {
	Role string `json:"role,omitempty"`
	Rank string `json:"rank,omitempty"`
}

type BanData struct {
	Banned      bool      `json:"banned"`
	Reason      string    `json:"reason"`
	Permanently bool      `json:"permanently"`
	Expiration  time.Time `json:"expiration"`
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
