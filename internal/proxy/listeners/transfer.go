package listeners

import (
	"context"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/edition/java/cookie"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

// send players to other proxies
func (lm *ListenerManager) onPreShutdown(e *proxy.PreShutdownEvent) {
	for _, p := range lm.ownerGate.Players() {
		proxy := lm.mm.GetProxyWithLowestPlayerCount(false)
		if proxy == nil {
			e.SetReason(util.TextError("The proxy you were on has closed and there was no other proxy to connect to."))
			continue
		}

		tr := lm.tm.BuildTask(tasks.NewTransferTask(p.ID(), lm.tm.GetOwnerId(), proxy.GetId(), uuid.Nil))
		if !tr.IsSuccessful() {
			lm.l.Error("transfer not successful", "error", tr.GetInfo())
			e.SetReason(util.TextError("The proxy you were on has closed and there was no other proxy to connect to."))
			continue
		}

		lm.l.Info("transferring player", "playerId", p.ID(), "proxyId", proxy.GetId())
	}
}

// check if player has cookie specifying which server he needs.
func (lm *ListenerManager) onChooseInitialServer(e *proxy.PlayerChooseInitialServerEvent) {
	p := e.Player()
	if len(lm.ownerGate.Servers()) < 1 {
		lm.l.Warn("no servers under gate proxy", "playerId", p.ID())
		lm.sendNoAvailableServers(p)
	} else {
		ctx, canc := context.WithTimeout(p.Context(), 2*time.Second)
		defer canc()

		c, err := cookie.Request(ctx, p, tasks.TransferKey, lm.ownerGate.Event())
		if err != nil {
			lm.l.Warn("transfer manager cookie request error", "error", err)
		}

		if err == nil && c != nil && len(c.Payload) > 0 {
			// reset
			err = cookie.Clear(p, tasks.TransferKey)
			if err != nil {
				lm.l.Error("transfer manager clearing cookie error", "error", err)
			}

			server_name := string(c.Payload)
			s := lm.ownerGate.Server(server_name)
			if s != nil {
				e.SetInitialServer(s)
			} else {
				lm.chooseRandomServer(p, e)
			}
		} else {
			lm.chooseRandomServer(p, e)
		}
	}
}

func (lm *ListenerManager) chooseRandomServer(p proxy.Player, e *proxy.PlayerChooseInitialServerEvent) {
	// for _, mb := range tm.mm.GetAllMultiBackendsUnderMultiProxy(tm.mm.GetOwnerMultiProxy()) {
	// 	//if util.IsBackendResponding(mb.GetAddress()) {
	// 	server := tm.p.Server(mb.GetId().String())
	// 	if server != nil {
	// 		l = append(l, server)
	// 	}
	// 	//}
	// }

	l := lm.ownerGate.Servers()

	if len(l) < 1 {
		lm.l.Warn("no servers under gate proxy", "playerId", p.ID())
		lm.sendNoAvailableServers(p)
		return
	}

	randomIndex := time.Now().UnixNano() % int64(len(l))
	e.SetInitialServer(l[randomIndex])
}

func (lm *ListenerManager) sendNoAvailableServers(player proxy.Player) {
	time.Sleep(50 * time.Millisecond)
	player.Disconnect(util.TextError("No available server. Please try again."))
}
