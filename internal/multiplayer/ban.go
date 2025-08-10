package multiplayer

import (
	"sync"
	"time"

	"go.minekube.com/common/minecraft/component"
)

type banInfo struct {
	banned bool
	reason component.Text

	permanently bool
	expiration  time.Duration

	mu sync.RWMutex
}

func newBanInfo() *banInfo {
	bi := &banInfo{
		banned:      false,
		reason:      component.Text{},
		permanently: false,
	}

	return bi
}

func (bi *banInfo) IsBanned() bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.banned
}

func (bi *banInfo) GetReason() component.Text {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.reason
}

func (bi *banInfo) Ban(reason component.Text) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	bi.banned = true
	bi.reason = reason
	bi.permanently = true
}

func (bi *banInfo) TempBan(reason component.Text, expiration time.Duration) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	bi.banned = true
	bi.reason = reason
	bi.permanently = false
	bi.expiration = expiration
}
