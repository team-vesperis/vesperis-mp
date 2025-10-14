package database

import (
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (db *Database) Test() error {
	id := uuid.New()

	data := &data.PlayerData{
		ProxyId:   uuid.Nil,
		BackendId: uuid.Nil,
		Username:  "User",
		Nickname:  "Bob",
		Permission: &data.PermissionData{
			Role: "default",
			Rank: "elite",
		},
		Ban: &data.BanData{
			Banned:      false,
			Reason:      "",
			Permanently: false,
			Expiration:  time.Time{},
		},
		Online:   false,
		Vanished: false,
		LastSeen: &time.Time{},
		Friends:  make([]uuid.UUID, 0),
	}

	err := db.SetPlayerData(id, data)
	if err != nil {
		db.l.Error("error setting test player data", "data", data, "error", err)
		return err
	}

	returnedData, err := db.GetPlayerData(id)
	if err != nil {
		db.l.Error("error getting test player data", "error", err)
		return err
	}

	db.l.Info("returned values", "nickname", returnedData.Nickname)

	var banned bool
	err = db.GetPlayerDataField(id, key.PlayerKey_Ban_Banned, &banned)
	if err != nil {
		db.l.Error("error getting test player data field", "error", err)
		return err
	}

	// this should happen
	if !banned {
		db.l.Info("successfully got banned")
	}

	return nil
}
