package multiplayer

import (
	"errors"
	"slices"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type PlayerPermissionManager struct {
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

const (
	RankChampion = "champion"
	RankDefault  = "default"
	RankElite    = "elite"
	RankLegend   = "legend"
)

var ErrIncorrectRank = errors.New("incorrect rank")

func InitPlayerPermissionManager(db *database.Database, l *logger.Logger) *PlayerPermissionManager {
	return &PlayerPermissionManager{
		db: db,
		l:  l,
	}
}

func (ppm *PlayerPermissionManager) GetRoleFromPlayer(p proxy.Player) (string, error) {
	return ppm.GetRoleFromId(p.ID().String())
}

func (ppm *PlayerPermissionManager) GetRoleFromId(id string) (string, error) {
	val, err := ppm.db.GetPlayerDataField(id, "permission.role")
	if err != nil {
		ppm.l.Warn("playerdata get role from database error", "playerId", id)
		return RoleDefault, err
	}

	role, ok := val.(string)
	if ok {
		if ppm.IsValidRole(role) {
			return role, nil
		}

		ppm.l.Warn("playerdata incorrect role returned", "playerId", id, "returnedRole", role)
		return RoleDefault, ErrIncorrectRole

	} else {
		ppm.l.Warn("playerdata incorrect value type returned from database", "playerId", id, "returnedValue", val)
		return RoleDefault, ErrIncorrectValueType
	}
}

func (ppm *PlayerPermissionManager) SetRoleWithPlayer(p proxy.Player, role string) error {
	return ppm.SetRoleWithId(p.ID().String(), role)
}

func (ppm *PlayerPermissionManager) SetRoleWithId(id string, role string) error {
	if !ppm.IsValidRole(role) {
		ppm.l.Warn("playerdata incorrect role returned", "playerId", id, "returnedRole", role)
		return ErrIncorrectRole
	}

	err := ppm.db.SetPlayerDataField(id, "permission.role", role)
	if err != nil {
		ppm.l.Warn("playerdata set role in database error", "playerId", id, "role", role)
	} else {
		ppm.l.Info("playerdata set new role", "playerId", id, "newRole", role)
	}

	return err
}

// Check if player has one of the following roles: admin, builder or moderator. Returns false if error occurs.
func (ppm *PlayerPermissionManager) IsPlayerPrivileged(p proxy.Player) bool {
	return ppm.IsIdPrivileged(p.ID().String())
}

// Check if id has one of the following roles: admin, builder or moderator. Returns false if error occurs.
func (ppm *PlayerPermissionManager) IsIdPrivileged(id string) bool {
	role, err := ppm.GetRoleFromId(id)
	if err != nil {
		return false
	}

	return ppm.IsRolePrivileged(role)
}

func (ppm *PlayerPermissionManager) IsRolePrivileged(role string) bool {
	return role == RoleAdmin || role == RoleBuilder || role == RoleModerator
}

func (ppm *PlayerPermissionManager) GetRankFromPlayer(p proxy.Player) (string, error) {
	return ppm.GetRankFromId(p.ID().String())
}

func (ppm *PlayerPermissionManager) GetRankFromId(id string) (string, error) {
	val, err := ppm.db.GetPlayerDataField(id, "permission.rank")
	if err != nil {
		ppm.l.Warn("playerdata get rank from database error", "playerId", id)
		return RankDefault, err
	}

	rank, ok := val.(string)
	if ok {
		if ppm.IsValidRank(rank) {
			return rank, nil
		}

		ppm.l.Warn("playerdata incorrect rank returned", "playerId", id, "returnedRank", rank)
		return RankDefault, ErrIncorrectRank

	} else {
		ppm.l.Warn("playerdata incorrect value type returned from database", "playerId", id, "returnedValue", val)
		return RankDefault, ErrIncorrectValueType
	}
}

func (ppm *PlayerPermissionManager) SetRankWithPlayer(p proxy.Player, rank string) error {
	return ppm.SetRankWithId(p.ID().String(), rank)
}

func (ppm *PlayerPermissionManager) SetRankWithId(id string, rank string) error {
	if !ppm.IsValidRank(rank) {
		ppm.l.Warn("playerdata incorrect rank returned", "playerId", id, "returnedRank", rank)
		return ErrIncorrectRank
	}

	err := ppm.db.SetPlayerDataField(id, "permission.rank", rank)
	if err != nil {
		ppm.l.Warn("playerdata set rank in database error", "playerId", id, "rank", rank)
	} else {
		ppm.l.Info("playerdata set new rank", "playerId", id, "newRank", rank)
	}

	return err
}

var validRoles = []string{
	RoleAdmin,
	RoleBuilder,
	RoleDefault,
	RoleModerator,
}

func (ppm *PlayerPermissionManager) IsValidRole(role string) bool {
	return slices.Contains(validRoles, role)
}

var validRanks = []string{
	RankChampion,
	RankDefault,
	RankElite,
	RankLegend,
}

func (ppm *PlayerPermissionManager) IsValidRank(rank string) bool {
	return slices.Contains(validRanks, rank)
}
