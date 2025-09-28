package multi

import (
	"sync"
	"time"
)

type banInfo struct {
	banned bool
	reason string

	permanently bool
	expiration  time.Duration

	mu sync.RWMutex

	mp *Player
}

func newBanInfo(mp *Player) *banInfo {
	bi := &banInfo{
		banned:      false,
		reason:      "",
		permanently: false,
		mp:          mp,
	}

	return bi
}

func (bi *banInfo) IsBanned() bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.banned
}

func (bi *banInfo) GetReason() string {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.reason
}

func (bi *banInfo) Ban(reason string) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	bi.banned = true
	bi.reason = reason
	bi.permanently = true
}

func (bi *banInfo) TempBan(reason string, expiration time.Duration) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	bi.banned = true
	bi.reason = reason
	bi.permanently = false
	bi.expiration = expiration
}
