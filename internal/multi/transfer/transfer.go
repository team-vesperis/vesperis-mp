package transfer

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/manager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/key"
	"go.minekube.com/gate/pkg/edition/java/cookie"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type TransferManager struct {
	l  *logger.Logger
	db *database.Database

	p  *proxy.Proxy
	mm *manager.MultiManager
}

var transferKey = key.New("vesperis", "transfer")

const transferRequestChannel = "transfer_request"

func Init(l *logger.Logger, db *database.Database, p *proxy.Proxy, mm *manager.MultiManager) *TransferManager {
	now := time.Now()
	tm := &TransferManager{
		l:  l,
		db: db,
		p:  p,
		mm: mm,
	}

	tm.listenToTransfers()
	tm.l.Info("initialized transfer manager", "duration", time.Since(now))
	return tm
}

func (tm *TransferManager) isBackendResponding(backend string) bool {
	conn, err := net.DialTimeout("tcp", backend, time.Second*5)
	if err == nil {
		conn.Close()
	}
	return err == nil
}

// send players to other proxies
func (tm *TransferManager) OnPreShutdown(event *proxy.PreShutdownEvent) {
	for _, p := range tm.p.Players() {
		proxy := tm.mm.GetProxyWithLowestPlayerCount(false)
		tm.l.Info("found a proxy to send a player to", "proxyId", proxy.GetId())
		err := tm.TransferPlayerToProxy(p, proxy.GetId())
		if err != nil {
			tm.disconnectPlayer(p)
		}

	}
}

func (tm *TransferManager) disconnectPlayer(p proxy.Player) {
	time.Sleep(50 * time.Millisecond)
	p.Disconnect(&component.Text{
		Content: "The proxy you were on has closed and there was no other proxy to connect to.",
		S:       util.StyleColorRed,
	})
}

// check if player has cookie specifying which server he needs.
func (tm *TransferManager) OnChooseInitialServer(event *proxy.PlayerChooseInitialServerEvent) {
	p := event.Player()
	if len(tm.p.Servers()) < 1 {
		tm.sendNoAvailableServers(p)
	} else {
		c, err := cookie.Request(p.Context(), p, transferKey, tm.p.Event())
		if err == nil && c != nil && len(c.Payload) > 0 {
			// reset
			err = cookie.Clear(p, transferKey)
			if err != nil {
				tm.l.Error("transfer manager clearing cookie error", "error", err)
			}

			server_name := string(c.Payload)
			s := tm.p.Server(server_name)
			if s != nil {
				event.SetInitialServer(s)
			} else {
				tm.chooseRandomServer(p, event)
			}
		} else {
			tm.chooseRandomServer(p, event)
		}
	}
}

func (tm *TransferManager) chooseRandomServer(player proxy.Player, event *proxy.PlayerChooseInitialServerEvent) {
	var servers []proxy.RegisteredServer
	for _, server := range tm.p.Servers() {
		if tm.isBackendResponding(server.ServerInfo().Addr().String()) {
			servers = append(servers, server)
		}
	}

	if len(servers) < 1 {
		tm.sendNoAvailableServers(player)
		return
	}

	randomIndex := time.Now().UnixNano() % int64(len(servers))
	event.SetInitialServer(servers[randomIndex])
}

func (tm *TransferManager) sendNoAvailableServers(player proxy.Player) {
	time.Sleep(50 * time.Millisecond)
	player.Disconnect(&component.Text{
		Content: "No available server. Please try again.",
		S: component.Style{
			Color: color.Red,
		},
	})
}

func (tm *TransferManager) TransferPlayerToServerOnOtherProxy(player proxy.Player, targetProxyId, targetBackendId uuid.UUID) error {
	responseChannel := "proxy_transfer_accept_" + player.ID().String()
	requestData := targetProxyId.String() + "|" + targetBackendId.String() + "|" + responseChannel
	msg, err := tm.db.SendAndReturn(transferRequestChannel, responseChannel, requestData, 2*time.Second)
	if err != nil {
		tm.l.Error("transfer manager send and return error", "error", err)
		return err
	}

	tm.l.Info(msg.Payload)

	parts := strings.Split(msg.Payload, "|")
	if len(parts) == 3 && parts[0] == player.ID().String() && parts[1] == targetProxyId.String() {
		// server will be one of four things:
		// 0, means the proxy is found but given server is not available
		// 1, means the proxy is found and none server is specified
		// 2, means the proxy is found and the given server is available
		// 3, means the proxy is found and the given server is found but not responding
		server := parts[3]

		if server == "0" {
			tm.l.Warn("Specified server for player transfer was not found: ", player.ID().String())
			return errors.New("specified server was not found")
		}

		if server == "3" {
			tm.l.Warn("Specified server for player transfer was found, but is not responding: ", player.ID().String())
			return errors.New("specified server found but not responding")
		}

		if server == "2" {
			c := &cookie.Cookie{
				Key:     transferKey,
				Payload: []byte(targetBackendId.String()),
			}

			err := cookie.Store(player, c)
			if err != nil {
				tm.l.Error("Error storing cookie to player: " + player.Username() + " - Error: " + err.Error())
				return errors.New("could not store cookie")
			}
		}

		address := parts[2]
		err := player.TransferToHost(address)
		if err != nil {
			return err
		}

		tm.l.Info("Player transfer successful: ", player.ID().String(), " to ", address)
	}

	return nil
}

func (tm *TransferManager) TransferPlayerToProxy(player proxy.Player, targetProxyId uuid.UUID) error {
	return tm.TransferPlayerToServerOnOtherProxy(player, targetProxyId, uuid.Nil)
}

func (tm *TransferManager) listenToTransfers() {
	tm.db.CreateListener(transferRequestChannel, func(msg *redis.Message) {
		parts := strings.Split(msg.Payload, "|")
		if len(parts) != 3 {
			tm.l.Error("Invalid transfer command format")
			return
		}

		tm.l.Info("retrieved transfer request")

		targetProxy := parts[1]
		if targetProxy == tm.mm.GetOwnerMultiProxy().GetId().String() {
			targetServer := parts[2]
			address := tm.p.Config().Bind
			responseChannel := parts[3]

			server := "0"
			if targetServer == "-" {
				server = "1"
			} else {
				foundServer := tm.p.Server(targetServer)
				if foundServer != nil {
					server = "2"
				}

				if !tm.isBackendResponding(foundServer.ServerInfo().Addr().String()) {
					server = "3"
				}
			}

			m := targetProxy + "|" + address + "|" + server
			err := tm.db.Publish(responseChannel, m)
			if err != nil {
				tm.l.Warn("transfer listener publish return error", "error", err)
			}
		}
	})
}
