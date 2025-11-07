package transfer

import (
	"context"
	"errors"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/manager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/key"
	"go.minekube.com/gate/pkg/edition/java/cookie"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type TransferManager struct {
	l  *logger.Logger
	db *database.Database

	tm *task.TaskManager
	p  *proxy.Proxy
	mm *manager.MultiManager
}

var transferKey = key.New("vesperis", "transfer")

func Init(l *logger.Logger, db *database.Database, tm *task.TaskManager, p *proxy.Proxy, mm *manager.MultiManager) *TransferManager {
	now := time.Now()
	trm := &TransferManager{
		l:  l,
		db: db,
		tm: tm,
		p:  p,
		mm: mm,
	}

	trm.l.Info("initialized transfer manager", "duration", time.Since(now))
	return trm
}

// send players to other proxies
func (tm *TransferManager) OnPreShutdown(e *proxy.PreShutdownEvent) {
	for _, p := range tm.p.Players() {
		proxy := tm.mm.GetProxyWithLowestPlayerCount(false)
		tm.l.Info("found a proxy to send a player to", "proxyId", proxy.GetId())
		err := tm.TransferPlayerToProxy(p, proxy)
		if err != nil {
			tm.disconnectPlayer(p)
		}

	}
}

// check if player has cookie specifying which server he needs.
func (tm *TransferManager) OnChooseInitialServer(e *proxy.PlayerChooseInitialServerEvent) {
	p := e.Player()
	if len(tm.p.Servers()) < 1 {
		tm.l.Warn("no servers under gate proxy", "playerId", p.ID())
		tm.sendNoAvailableServers(p)
	} else {
		ctx, canc := context.WithTimeout(p.Context(), 2*time.Second)
		defer canc()

		c, err := cookie.Request(ctx, p, transferKey, tm.p.Event())
		if err != nil {
			tm.l.Warn("transfer manager cookie request error", "error", err)
		}

		if err == nil && c != nil && len(c.Payload) > 0 {
			// reset
			err = cookie.Clear(p, transferKey)
			if err != nil {
				tm.l.Error("transfer manager clearing cookie error", "error", err)
			}

			server_name := string(c.Payload)
			s := tm.p.Server(server_name)
			if s != nil {
				e.SetInitialServer(s)
			} else {
				tm.chooseRandomServer(p, e)
			}
		} else {
			tm.chooseRandomServer(p, e)
		}
	}
}

func (tm *TransferManager) chooseRandomServer(p proxy.Player, e *proxy.PlayerChooseInitialServerEvent) {
	// for _, mb := range tm.mm.GetAllMultiBackendsUnderMultiProxy(tm.mm.GetOwnerMultiProxy()) {
	// 	//if util.IsBackendResponding(mb.GetAddress()) {
	// 	server := tm.p.Server(mb.GetId().String())
	// 	if server != nil {
	// 		l = append(l, server)
	// 	}
	// 	//}
	// }

	l := tm.p.Servers()

	if len(l) < 1 {
		tm.l.Warn("no servers under gate proxy", "playerId", p.ID())
		tm.sendNoAvailableServers(p)
		return
	}

	randomIndex := time.Now().UnixNano() % int64(len(l))
	e.SetInitialServer(l[randomIndex])
}

func (tm *TransferManager) disconnectPlayer(p proxy.Player) {
	time.Sleep(50 * time.Millisecond)
	p.Disconnect(&c.Text{
		Content: "The proxy you were on has closed and there was no other proxy to connect to.",
		S:       util.StyleColorRed,
	})
}

func (tm *TransferManager) sendNoAvailableServers(player proxy.Player) {
	time.Sleep(50 * time.Millisecond)
	player.Disconnect(&c.Text{
		Content: "No available server. Please try again.",
		S: c.Style{
			Color: color.Red,
		},
	})
}

var (
	ErrSpecifiedServerNotFound              = errors.New("specified server was not found")
	ErrSpecifiedServerFoundButNotResponding = errors.New("specified server was found but not responding")
)

func (tm *TransferManager) TransferPlayerToServerOnOtherProxy(p proxy.Player, mp *multi.Proxy, targetBackendId uuid.UUID) error {
	tr := tm.tm.BuildTask(tasks.NewTransferRequestTask(mp.GetId(), targetBackendId))
	if !tr.IsSuccessful() {
		return errors.New(tr.GetInfo())
	}

	// tr.GetInfo() will be one of four things:
	// 0, given server is not available
	// 1, given server is found but not responding
	// 2, given server is available
	// 3, none server is specified

	if tr.GetInfo() == "0" {
		tm.l.Warn("transfer manager specified server not found error", "playerId", p.ID(), "targetBackendId", targetBackendId)
		return ErrSpecifiedServerNotFound
	}

	if tr.GetInfo() == "1" {
		tm.l.Warn("transfer manager specified server found but not responding error", "playerId", p.ID(), "targetBackendId", targetBackendId)
		return ErrSpecifiedServerFoundButNotResponding
	}

	if tr.GetInfo() == "2" {
		c := &cookie.Cookie{
			Key:     transferKey,
			Payload: []byte(targetBackendId.String()),
		}

		err := cookie.Store(p, c)
		if err != nil {
			tm.l.Warn("transfer manager could not store cookie on player", "playerId", p.ID(), "cookie", c)
			return err
		}
	}

	err := p.TransferToHost(mp.GetAddress())
	if err != nil {
		return err
	}

	tm.l.Info("player transfer successful", "playerId", p.ID(), "proxyId", mp.GetId())

	return nil
}

func (tm *TransferManager) TransferPlayerToProxy(p proxy.Player, mp *multi.Proxy) error {
	return tm.TransferPlayerToServerOnOtherProxy(p, mp, uuid.Nil)
}
