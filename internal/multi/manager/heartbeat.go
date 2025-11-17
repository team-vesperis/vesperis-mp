package manager

import (
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
)

type hartBeatManager struct {
	t  *time.Ticker
	p  *multi.Proxy
	d  chan bool
	mm *MultiManager
}

const maxLastHartBeat = 5 * time.Minute

func (mm *MultiManager) InitHeartBeatManager() *hartBeatManager {
	hbm := &hartBeatManager{
		t:  time.NewTicker(3 * time.Minute),
		p:  mm.ownerMP,
		d:  make(chan bool),
		mm: mm,
	}

	mm.hbm = hbm
	go hbm.start()

	return hbm
}

func (hbm *hartBeatManager) start() {
	for {
		select {
		case <-hbm.d:
			return
		case t := <-hbm.t.C:
			hbm.p.SetLastHeartBeat(&t)
			go func() {
				now := time.Now()
				lockKey := "proxy_cleanup_leader"
				got, err := hbm.mm.GetDatabase().AcquireLock(lockKey, 30*time.Second)
				if err != nil {
					hbm.mm.l.Warn("could not acquire cleanup leader lock", "error", err)
					return
				}
				if !got {
					return
				}

				hbm.checkOtherProxies()
				d := time.Since(now)
				time.Sleep((3 * time.Minute) - d)
				hbm.mm.GetDatabase().ReleaseLock(lockKey)
			}()
		}
	}
}

func (hbm *hartBeatManager) stop() {
	hbm.t.Stop()
	hbm.d <- true
}

func (hbm *hartBeatManager) checkOtherProxies() {
	now := time.Now()

	for _, mp := range hbm.mm.GetAllMultiProxies() {
		if mp == hbm.mm.GetOwnerMultiProxy() {
			continue
		}

		lhb := mp.GetLastHeartBeat()
		if lhb == nil || now.Sub(*lhb) > maxLastHartBeat {
			for _, b_id := range mp.GetBackendsIds() {
				err := hbm.mm.DeleteMultiBackend(b_id)
				if err != nil {
					hbm.mm.l.Warn("hart beat manager delete multibackend error", "backendId", b_id, "error", err)
				}
			}

			for _, p_id := range mp.GetPlayerIds() {
				p, err := hbm.mm.GetMultiPlayer(p_id)
				if err != nil {
					hbm.mm.l.Warn("hart beat manager delete multiplayer error", "playerId", p_id, "error", err)
					continue
				}

				err = p.SetProxy(nil)
				if err != nil {
					hbm.mm.l.Warn("hart beat manager set multiplayer's proxy error", "playerId", p_id, "error", err)
				}

				if p.IsOnline() {
					err := p.SetOnline(false)
					if err != nil {
						hbm.mm.l.Warn("hart beat manager set multiplayer's online error", "playerId", p_id, "error", err)
					}
				}

				err = p.SetLastSeen(&now)
				if err != nil {
					hbm.mm.l.Warn("hart beat manager set multiplayer's last seen", "playerId", p_id, "error", err)
				}
			}

			err := hbm.mm.DeleteMultiProxy(mp.GetId())
			if err != nil {
				hbm.mm.l.Error("hart beat manager delete multiproxy error", "proxyId", mp.GetId(), "error", err)
			}
		}
	}
}
