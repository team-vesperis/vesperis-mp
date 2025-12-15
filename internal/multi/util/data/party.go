package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"go.minekube.com/gate/pkg/util/uuid"
)

type PartyData struct {
	// The playerId of the owner of the partner. The owner can invite/allow/remove other players.
	PartyOwner uuid.UUID `json:"partyOwner"`

	// List of playerIds that are members of the party.
	PartyMembers []uuid.UUID `json:"partyMembers"`

	// List of playerIds that are invited to the party.
	PartyInvitations []uuid.UUID `json:"partyInvitations"`

	// List of playerIds that are requesting to join the party.
	PartyJoinRequests []uuid.UUID `json:"partyJoinRequests"`
}

func (pd PartyData) Value() (driver.Value, error) {
	return json.Marshal(pd)
}

func (pd *PartyData) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, pd)
	case string:
		return json.Unmarshal([]byte(v), pd)
	case nil:
		*pd = PartyData{}
		return nil
	default:
		return errors.New("unsupported type for PartyData Scan")
	}
}
