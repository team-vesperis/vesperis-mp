package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"go.minekube.com/gate/pkg/util/uuid"
)

type PlayerData struct {
	Proxy   uuid.UUID `json:"proxy"`
	Backend uuid.UUID `json:"backend"`

	Username string `json:"username"`
	Nickname string `json:"nickname"`

	Permission *PermissionData `json:"permission"`
	Ban        *BanData        `json:"ban"`
	Friend     *FriendData     `json:"friend"`

	// The party id the player is in. If not in party uuid.Nil
	PartyId uuid.UUID `json:"partyId"`
	// List of party ids where the player is invited to.
	PartyInvitations []uuid.UUID `json:"partyInvitations"`

	Online   bool       `json:"online"`
	Vanished bool       `json:"vanished"`
	LastSeen *time.Time `json:"lastSeen"`
}

type FriendData struct {
	Friends               []uuid.UUID `json:"friends"`
	FriendRequests        []uuid.UUID `json:"friendRequests"`
	FriendPendingRequests []uuid.UUID `json:"friendPendingRequests"`
}

type PermissionData struct {
	Role string `json:"role"`
	Rank string `json:"rank"`
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

func (pd *PlayerData) InitializeDefaults() bool {
	initialized := false

	if pd.Permission == nil {
		pd.Permission = &PermissionData{
			Role: "default",
			Rank: "default",
		}
		initialized = true
	}

	if pd.Ban == nil {
		pd.Ban = &BanData{
			Banned:      false,
			Reason:      "",
			Permanently: false,
			Expiration:  time.Time{},
		}
		initialized = true
	}

	if pd.Friend == nil {
		pd.Friend = &FriendData{
			Friends:               make([]uuid.UUID, 0),
			FriendRequests:        make([]uuid.UUID, 0),
			FriendPendingRequests: make([]uuid.UUID, 0),
		}
		initialized = true
	}

	return initialized
}
