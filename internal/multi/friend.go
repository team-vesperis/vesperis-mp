package multi

import (
	"errors"
	"slices"
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

type friendInfo struct {
	friends               []uuid.UUID
	friendRequests        []uuid.UUID
	friendPendingRequests []uuid.UUID

	mu sync.RWMutex
	mp *Player
}

func newFriendInfo(mp *Player, data *data.PlayerData) *friendInfo {
	fi := &friendInfo{
		friends:               data.Friend.Friends,
		friendRequests:        data.Friend.FriendRequests,
		friendPendingRequests: data.Friend.FriendPendingRequests,

		mu: sync.RWMutex{},
		mp: mp,
	}

	return fi
}

var ErrFriendNotFound = errors.New("friend not found")

func (fi *friendInfo) GetFriendRequestIds() []uuid.UUID {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	return slices.Clone(fi.friendRequests)
}

func (fi *friendInfo) SetFriendRequestIds(ids []uuid.UUID) error {
	return fi.setFriendRequestIds(ids, true)
}

func (fi *friendInfo) setFriendRequestIds(ids []uuid.UUID, notify bool) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.friendRequests = ids

	if notify {
		return fi.mp.save(key.PlayerKey_Friend_FriendRequests, ids)
	}

	return nil
}

func (fi *friendInfo) AddFriendRequestId(id uuid.UUID) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	if !slices.Contains(fi.friendRequests, id) {
		fi.friendRequests = append(fi.friendRequests, id)
	}

	return fi.mp.save(key.PlayerKey_Friend_FriendRequests, fi.friendRequests)
}

func (fi *friendInfo) RemoveFriendRequestId(id uuid.UUID) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	i := slices.Index(fi.friendRequests, id)
	if i == -1 {
		return ErrFriendNotFound
	}
	fi.friendRequests = slices.Delete(fi.friendRequests, i, i+1)

	return fi.mp.save(key.PlayerKey_Friend_FriendRequests, fi.friendRequests)
}

func (fi *friendInfo) GetPendingFriendRequestIds() []uuid.UUID {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	return slices.Clone(fi.friendPendingRequests)
}

func (fi *friendInfo) SetPendingFriendIds(ids []uuid.UUID) error {
	return fi.setPendingFriendIds(ids, true)
}

func (fi *friendInfo) setPendingFriendIds(ids []uuid.UUID, notify bool) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.friendPendingRequests = ids

	if notify {
		return fi.mp.save(key.PlayerKey_Friend_FriendPendingRequests, ids)
	}

	return nil
}

func (fi *friendInfo) AddPendingFriendRequestId(id uuid.UUID) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	if !slices.Contains(fi.friendPendingRequests, id) {
		fi.friendPendingRequests = append(fi.friendPendingRequests, id)
	}

	return fi.mp.save(key.PlayerKey_Friend_FriendPendingRequests, fi.friendPendingRequests)
}

func (fi *friendInfo) RemovePendingFriendRequestId(id uuid.UUID) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	i := slices.Index(fi.friendPendingRequests, id)
	if i == -1 {
		return ErrFriendNotFound
	}
	fi.friendPendingRequests = slices.Delete(fi.friendPendingRequests, i, i+1)

	return fi.mp.save(key.PlayerKey_Friend_FriendPendingRequests, fi.friendPendingRequests)
}

func (fi *friendInfo) GetFriendsIds() []uuid.UUID {
	fi.mu.RLock()
	defer fi.mu.RUnlock()

	return slices.Clone(fi.friends)
}

func (fi *friendInfo) SetFriendsIds(ids []uuid.UUID) error {
	return fi.setFriendsIds(ids, true)
}

func (fi *friendInfo) setFriendsIds(ids []uuid.UUID, notify bool) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.friends = ids

	if notify {
		return fi.mp.save(key.PlayerKey_Friend_Friends, ids)
	}

	return nil
}

func (fi *friendInfo) AddFriendId(id uuid.UUID) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	if !slices.Contains(fi.friends, id) {
		fi.friends = append(fi.friends, id)
	}

	return fi.mp.save(key.PlayerKey_Friend_Friends, fi.friends)
}

func (fi *friendInfo) RemoveFriendId(id uuid.UUID) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	i := slices.Index(fi.friends, id)
	if i == -1 {
		return ErrFriendNotFound
	}
	fi.friends = slices.Delete(fi.friends, i, i+1)

	return fi.mp.save(key.PlayerKey_Friend_Friends, fi.friends)

}
