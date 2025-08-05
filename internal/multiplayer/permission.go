package multiplayer

import (
	"errors"
	"slices"
)

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

var validRoles = []string{
	RoleAdmin,
	RoleBuilder,
	RoleDefault,
	RoleModerator,
}

func IsValidRole(role string) bool {
	return slices.Contains(validRoles, role)
}

var validRanks = []string{
	RankChampion,
	RankDefault,
	RankElite,
	RankLegend,
}

func IsValidRank(rank string) bool {
	return slices.Contains(validRanks, rank)
}
