package multi

import (
	"sync"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
)

type banInfo struct {
	banned bool
	reason string

	permanently bool
	expiration  time.Time

	mu sync.RWMutex
	mp *Player
}

func newBanInfo(mp *Player, data *data.PlayerData) *banInfo {
	bi := &banInfo{
		banned:      data.Ban.Banned,
		reason:      data.Ban.Reason,
		permanently: data.Ban.Permanently,
		expiration:  data.Ban.Expiration,

		mp: mp,
		mu: sync.RWMutex{},
	}

	return bi
}

func (bi *banInfo) IsBanned() bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.banned
}

func (bi *banInfo) setBanned(banned bool, notify bool) error {
	bi.mu.Lock()
	bi.banned = banned
	bi.mu.Unlock()

	if notify {
		return bi.mp.save(key.PlayerKey_Ban_Banned, banned)
	}

	return nil
}

func (bi *banInfo) GetReason() string {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.reason
}

func (bi *banInfo) setReason(reason string, notify bool) error {
	bi.mu.Lock()
	bi.reason = reason
	bi.mu.Unlock()

	if notify {
		return bi.mp.save(key.PlayerKey_Ban_Reason, reason)
	}

	return nil
}

func (bi *banInfo) IsPermanently() bool {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.permanently
}

func (bi *banInfo) setPermanently(permanently bool, notify bool) error {
	bi.mu.Lock()
	bi.permanently = permanently
	bi.mu.Unlock()

	if notify {
		return bi.mp.save(key.PlayerKey_Ban_Permanently, permanently)
	}

	return nil
}

func (bi *banInfo) GetExpiration() time.Time {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	return bi.expiration
}

func (bi *banInfo) setExpiration(expiration time.Time, notify bool) error {
	bi.mu.Lock()
	bi.expiration = expiration
	bi.mu.Unlock()

	if notify {
		return bi.mp.save(key.PlayerKey_Ban_Expiration, expiration)
	}

	return nil
}

func (bi *banInfo) Ban(reason string) error {
	err := bi.setBanned(true, true)
	if err != nil {
		return err
	}

	err = bi.setReason(reason, true)
	if err != nil {
		return err
	}

	err = bi.setPermanently(true, true)
	if err != nil {
		return err
	}

	return bi.setExpiration(time.Time{}, true)
}

func (bi *banInfo) TempBan(reason string, expiration time.Time) error {
	err := bi.setBanned(true, true)
	if err != nil {
		return err
	}

	err = bi.setReason(reason, true)
	if err != nil {
		return err
	}

	err = bi.setPermanently(false, true)
	if err != nil {
		return err
	}

	return bi.setExpiration(expiration, true)
}

func (bi *banInfo) UnBan() error {
	err := bi.setBanned(false, true)
	if err != nil {
		return err
	}

	err = bi.setReason("", true)
	if err != nil {
		return err
	}

	err = bi.setPermanently(false, true)
	if err != nil {
		return err
	}

	return bi.setExpiration(time.Time{}, true)
}
