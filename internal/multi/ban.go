package multi

import (
	"sync"
	"time"

	c "go.minekube.com/common/minecraft/component"
)

type banInfo struct {
	banned bool
	reason c.Text

	permanently bool
	expiration  time.Duration

	mu sync.RWMutex

	mp *Player
}

func newBanInfo(mp *Player) *banInfo {
	bi := &banInfo{
		banned:      false,
		reason:      c.Text{},
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

func (bi *banInfo) GetReason() c.Text {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.reason
}

func (bi *banInfo) Ban(reason c.Text) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	bi.banned = true
	bi.reason = reason
	bi.permanently = true
}

func (bi *banInfo) TempBan(reason c.Text, expiration time.Duration) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	bi.banned = true
	bi.reason = reason
	bi.permanently = false
	bi.expiration = expiration
}
