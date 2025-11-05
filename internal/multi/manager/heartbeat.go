package manager

import (
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi"
)

type hartBeatManager struct {
	t *time.Ticker
	p *multi.Proxy
	d chan bool
}

func (mm *MultiManager) InitHeartBeatManager() *hartBeatManager {
	hbm := &hartBeatManager{
		t: time.NewTicker(3 * time.Minute),
		p: mm.ownerMP,
		d: make(chan bool),
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
		}
	}
}

func (hbm *hartBeatManager) stop() {
	hbm.t.Stop()
	hbm.d <- true
}
