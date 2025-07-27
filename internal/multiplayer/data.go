package multiplayer

import (
	"errors"
	"slices"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type PlayerDataManager struct {
	db *database.Database
	l  *logger.Logger
}

var ErrIncorrectValueType = errors.New("incorrect value type returned from database")

const (
	RoleAdmin     = "admin"
	RoleBuilder   = "builder"
	RoleDefault   = "default"
	RoleModerator = "moderator"
)

var ErrIncorrectRole = errors.New("incorrect role")

var validRoles = []string{
	RoleAdmin,
	RoleBuilder,
	RoleDefault,
	RoleModerator,
}

var ErrIncorrectRank = errors.New("incorrect rank")

const (
	RankChampion = "champion"
	RankDefault  = "default"
	RankElite    = "elite"
	RankLegend   = "legend"
)

var validRanks = []string{
	RankChampion,
	RankDefault,
	RankElite,
	RankLegend,
}

func InitPlayerDataManager(db *database.Database, l *logger.Logger) *PlayerDataManager {
	return &PlayerDataManager{
		db: db,
		l:  l,
	}
}

func (pdm *PlayerDataManager) IsPlayerVanished(p proxy.Player) (bool, error) {
	return pdm.IsIdVanished(p.ID().String())
}

func (pdm *PlayerDataManager) IsIdVanished(id string) (bool, error) {
	val, err := pdm.db.GetPlayerDataField(id, "vanished")
	if err != nil {
		pdm.l.Warn("playerdata get vanished from database error", "playerId", id)
		return false, err
	}

	vanished, ok := val.(bool)
	if ok {
		return vanished, nil
	} else {
		pdm.l.Warn("playerdata incorrect value type returned from database", "playerId", id, "returnedValue", val)
		return false, ErrIncorrectValueType
	}
}

func (pdm *PlayerDataManager) SetPlayerVanished(p proxy.Player, vanished bool) error {
	return pdm.SetIdVanished(p.ID().String(), vanished)
}

func (pdm *PlayerDataManager) SetIdVanished(id string, vanished bool) error {
	err := pdm.db.SetPlayerDataField(id, "vanished", vanished)
	if err != nil {
		pdm.l.Warn("playerdata set vanished in database error", "playerId", id, "vanished", vanished)
	}

	return err
}

func (pdm *PlayerDataManager) GetRoleFromPlayer(p proxy.Player) (string, error) {
	return pdm.GetRoleFromId(p.ID().String())
}

func (pdm *PlayerDataManager) GetRoleFromId(id string) (string, error) {
	val, err := pdm.db.GetPlayerDataField(id, "role")
	if err != nil {
		pdm.l.Warn("playerdata get role from database error", "playerId", id)
		return "", err
	}

	role, ok := val.(string)
	if ok {
		if slices.Contains(validRoles, role) {
			return role, nil
		}

		pdm.l.Warn("playerdata incorrect role returned", "playerId", id, "returnedRole", role)
		return RoleDefault, ErrIncorrectRole

	} else {
		pdm.l.Warn("playerdata incorrect value type returned from database", "playerId", id, "returnedValue", val)
		return RoleDefault, ErrIncorrectValueType
	}
}

func (pdm *PlayerDataManager) SetRoleWithPlayer(p proxy.Player, role string) error {
	return pdm.SetRoleWithId(p.ID().String(), role)
}

func (pdm *PlayerDataManager) SetRoleWithId(id string, role string) error {
	if !slices.Contains(validRoles, role) {
		pdm.l.Warn("playerdata incorrect role returned", "playerId", id, "returnedRole", role)
		return ErrIncorrectRole
	}

	err := pdm.db.SetPlayerDataField(id, "role", role)
	if err != nil {
		pdm.l.Warn("playerdata set role in database error", "playerId", id, "role", role)
	} else {
		pdm.l.Info("playerdata set new role", "playerId", id, "newRole", role)
	}

	return err
}

// Check if player has one of the following roles: admin, builder or moderator. Returns false if error occurs.
func (pdm *PlayerDataManager) IsPlayerPrivileged(p proxy.Player) bool {
	return pdm.IsIdPrivileged(p.ID().String())
}

// Check if id has one of the following roles: admin, builder or moderator. Returns false if error occurs.
func (pdm *PlayerDataManager) IsIdPrivileged(id string) bool {
	role, err := pdm.GetRoleFromId(id)
	if err != nil {
		return false
	}

	return pdm.IsRolePrivileged(role)
}

func (pdm *PlayerDataManager) IsRolePrivileged(role string) bool {
	return role == RoleAdmin || role == RoleBuilder || role == RoleModerator
}

func (pdm *PlayerDataManager) GetRankFromPlayer(p proxy.Player) (string, error) {
	return pdm.GetRankFromId(p.ID().String())
}

func (pdm *PlayerDataManager) GetRankFromId(id string) (string, error) {
	val, err := pdm.db.GetPlayerDataField(id, "rank")
	if err != nil {
		pdm.l.Warn("playerdata get rank from database error", "playerId", id)
		return "", err
	}

	rank, ok := val.(string)
	if ok {
		if slices.Contains(validRanks, rank) {
			return rank, nil
		}

		pdm.l.Warn("playerdata incorrect rank returned", "playerId", id, "returnedRank", rank)
		return RankDefault, ErrIncorrectRank

	} else {
		pdm.l.Warn("playerdata incorrect value type returned from database", "playerId", id, "returnedValue", val)
		return RankDefault, ErrIncorrectValueType
	}
}

func (pdm *PlayerDataManager) SetRankWithPlayer(p proxy.Player, rank string) error {
	return pdm.SetRankWithId(p.ID().String(), rank)
}

func (pdm *PlayerDataManager) SetRankWithId(id string, rank string) error {
	if !slices.Contains(validRanks, rank) {
		pdm.l.Warn("playerdata incorrect rank returned", "playerId", id, "returnedRank", rank)
		return ErrIncorrectRank
	}

	err := pdm.db.SetPlayerDataField(id, "rank", rank)
	if err != nil {
		pdm.l.Warn("playerdata set rank in database error", "playerId", id, "rank", rank)
	} else {
		pdm.l.Info("playerdata set new rank", "playerId", id, "newRank", rank)
	}

	return err
}
