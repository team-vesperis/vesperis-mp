package multiplayer

import (
	"errors"
	"slices"
	"sync"
)

type permissionInfo struct {
	role string
	rank string

	mu sync.RWMutex

	mp *MultiPlayer
}

func newPermissionInfo(mp *MultiPlayer) *permissionInfo {
	pi := &permissionInfo{
		role: RoleDefault,
		rank: RankDefault,
		mp:   mp,
	}

	return pi
}

func (pi *permissionInfo) GetRole() string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.role
}

func (pi *permissionInfo) SetRole(role string, notify bool) error {
	if !IsValidRole(role) {
		return ErrIncorrectRole
	}

	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.role = role

	var err error
	if notify {
		err = pi.mp.save("permission.role", role)
	}

	return err
}

func (pi *permissionInfo) IsPrivileged() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.role == RoleAdmin || pi.role == RoleBuilder || pi.role == RoleModerator
}

func (pi *permissionInfo) GetRank() string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.rank
}

func (pi *permissionInfo) SetRank(rank string, notify bool) error {
	if !IsValidRank(rank) {
		return ErrIncorrectRank
	}

	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.rank = rank

	var err error
	if notify {
		err = pi.mp.save("permission.rank", rank)
	}

	return err
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
