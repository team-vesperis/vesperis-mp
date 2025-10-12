package multi

import (
	"errors"
	"slices"
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
)

type permissionInfo struct {
	role string
	rank string

	mu sync.RWMutex

	mp *Player
}

func newPermissionInfo(mp *Player, data *util.PlayerData) *permissionInfo {
	pi := &permissionInfo{
		role: data.Permission.Role,
		rank: data.Permission.Rank,
		mp:   mp,
	}

	return pi
}

func (pi *permissionInfo) GetRole() string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.role
}

func (pi *permissionInfo) SetRole(role string) error {
	return pi.setRole(role, true)
}

func (pi *permissionInfo) setRole(role string, notify bool) error {
	if !IsValidRole(role) {
		return ErrIncorrectRole
	}

	pi.mu.Lock()
	pi.role = role
	pi.mu.Unlock()

	if notify {
		return pi.mp.save(util.PlayerKey_Permission_Role, role)
	}

	return nil
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

func (pi *permissionInfo) SetRank(rank string) error {
	if !IsValidRank(rank) {
		return ErrIncorrectRank
	}

	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.rank = rank

	return pi.mp.save("permission.rank", rank)
}

func (pi *permissionInfo) setRank(rank string, notify bool) error {
	if !IsValidRank(rank) {
		return ErrIncorrectRank
	}

	pi.mu.Lock()
	pi.rank = rank
	pi.mu.Unlock()

	if notify {
		return pi.mp.save(util.PlayerKey_Permission_Rank, rank)
	}

	return nil
}

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
