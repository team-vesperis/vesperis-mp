package multi

import (
	"errors"
	"slices"
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
)

type permissionInfo struct {
	role Role
	rank Rank

	mu sync.RWMutex

	mp *Player
}

func newPermissionInfo(mp *Player, data *data.PlayerData) *permissionInfo {
	role, _ := GetRole(data.Permission.Role)
	rank, _ := GetRank(data.Permission.Rank)

	pi := &permissionInfo{
		role: role,
		rank: rank,
		mp:   mp,
		mu:   sync.RWMutex{},
	}

	return pi
}

func (pi *permissionInfo) GetRole() Role {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.role
}

func (pi *permissionInfo) SetRole(role Role) error {
	return pi.setRole(role, true)
}

func (pi *permissionInfo) setRole(role Role, notify bool) error {
	if !IsValidRole(role) {
		return ErrIncorrectRole
	}

	pi.mu.Lock()
	pi.role = role
	pi.mu.Unlock()

	if notify {
		return pi.mp.save(key.PlayerKey_Permission_Role, role)
	}

	return nil
}

func (pi *permissionInfo) IsPrivileged() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.role == RoleAdmin || pi.role == RoleBuilder || pi.role == RoleModerator
}

func (pi *permissionInfo) GetRank() Rank {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return pi.rank
}

func (pi *permissionInfo) SetRank(rank Rank) error {
	if !IsValidRank(rank) {
		return ErrIncorrectRank
	}

	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.rank = rank

	return pi.mp.save("permission.rank", rank)
}

func (pi *permissionInfo) setRank(rank Rank, notify bool) error {
	if !IsValidRank(rank) {
		return ErrIncorrectRank
	}

	pi.mu.Lock()
	pi.rank = rank
	pi.mu.Unlock()

	if notify {
		return pi.mp.save(key.PlayerKey_Permission_Rank, rank)
	}

	return nil
}

type Role string

func (r Role) String() string {
	return string(r)
}

func GetRole(s string) (Role, error) {
	r := Role(s)
	if !IsValidRole(r) {
		return Role(""), ErrIncorrectRole
	}

	return r, nil
}

const (
	RoleAdmin     Role = "admin"
	RoleBuilder   Role = "builder"
	RoleDefault   Role = "default"
	RoleModerator Role = "moderator"
)

var ErrIncorrectRole = errors.New("incorrect role")

type Rank string

func (r Rank) String() string {
	return string(r)
}

func GetRank(s string) (Rank, error) {
	r := Rank(s)
	if !IsValidRank(r) {
		return Rank(""), ErrIncorrectRank
	}

	return r, nil
}

const (
	RankChampion Rank = "champion"
	RankDefault  Rank = "default"
	RankElite    Rank = "elite"
	RankLegend   Rank = "legend"
)

var ErrIncorrectRank = errors.New("incorrect rank")

var validRoles = []Role{
	RoleAdmin,
	RoleBuilder,
	RoleDefault,
	RoleModerator,
}

func IsValidRole(r Role) bool {
	return slices.Contains(validRoles, r)
}

var validRanks = []Rank{
	RankChampion,
	RankDefault,
	RankElite,
	RankLegend,
}

func IsValidRank(r Rank) bool {
	return slices.Contains(validRanks, r)
}
