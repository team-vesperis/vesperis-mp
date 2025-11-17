package database

import (
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (db *Database) Test() error {
	for range 10000 {
		id := uuid.New()

		data := &data.PlayerData{
			ProxyId:   uuid.Nil,
			BackendId: uuid.Nil,
			Username:  "User-" + id.Undashed(),
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
	}

	return nil
}
