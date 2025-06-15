package transfer

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/mp/database"
	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/key"
	"go.minekube.com/gate/pkg/edition/java/cookie"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p           *proxy.Proxy
	logger      *zap.SugaredLogger
	proxy_name  string
	transferKey = key.New("vesperis", "transfer_specific_server")
)

func InitializeTransfer(proxy *proxy.Proxy, log *zap.SugaredLogger, pn string) {
	p = proxy
	logger = log
	proxy_name = pn
	go listenToTransfers()
	logger.Info("Initialized transfer.")
}

// send players to other proxies
func OnPreShutdown(event *proxy.PreShutdownEvent) {
	for _, player := range p.Players() {
		proxy, err := datasync.GetProxyWithLowestPlayerCount(false)
		logger.Info(proxy)
		if err != nil {
			logger.Error("Error getting the proxy with lowest player count for transfer: ", err)
			time.Sleep(50 * time.Millisecond)
			player.Disconnect(&component.Text{
				Content: "The proxy you were on has closed and there was no other proxy to connect to.",
				S: component.Style{
					Color: color.Red,
				},
			})
		} else {
			err = TransferPlayerToProxy(player, proxy)
			if err != nil {
				time.Sleep(50 * time.Millisecond)
				player.Disconnect(&component.Text{
					Content: "The proxy you were on has closed and there was no other proxy to connect to.",
					S: component.Style{
						Color: color.Red,
					},
				})
			}
		}
	}
}

func isBackendResponding(backend string) bool {
	conn, err := net.DialTimeout("tcp", backend, time.Second*5)
	if err == nil {
		conn.Close()
	}
	return err == nil
}

// check if player has cookie specifying which server he needs.
func OnChooseInitialServer(event *proxy.PlayerChooseInitialServerEvent) {
	player := event.Player()
	if len(p.Servers()) < 1 {
		sendNoAvailableServers(player)
	} else {
		c, err := cookie.Request(player.Context(), player, transferKey, p.Event())
		if err == nil && c != nil && len(c.Payload) > 0 {
			// reset
			err = cookie.Clear(player, transferKey)
			if err != nil {
				logger.Error("Error clearing cookie: " + err.Error())
			}

			server_name := string(c.Payload)
			server := p.Server(server_name)
			if server != nil {
				event.SetInitialServer(server)
			} else {
				chooseRandomServer(player, event)
			}
		} else {
			chooseRandomServer(player, event)
		}
	}
}

func chooseRandomServer(player proxy.Player, event *proxy.PlayerChooseInitialServerEvent) {
	var servers []proxy.RegisteredServer
	for _, server := range p.Servers() {
		if isBackendResponding(server.ServerInfo().Addr().String()) {
			servers = append(servers, server)
		}
	}

	if len(servers) < 1 {
		sendNoAvailableServers(player)
		return
	}

	randomIndex := time.Now().UnixNano() % int64(len(servers))
	event.SetInitialServer(servers[randomIndex])
}

func sendNoAvailableServers(player proxy.Player) {
	time.Sleep(50 * time.Millisecond)
	player.Disconnect(&component.Text{
		Content: "No available server. Please try again.",
		S: component.Style{
			Color: color.Red,
		},
	})
}

func TransferPlayerToServerOnOtherProxy(player proxy.Player, targetProxy string, targetServer string) error {
	responseChannel := "proxy_transfer_accept_" + player.ID().String()
	logger.Info(responseChannel)

	requestData := player.ID().String() + "|" + targetProxy + "|" + targetServer + "|" + responseChannel
	msg, err := database.SendAndReturn(context.Background(), "proxy_transfer_request", responseChannel, requestData, 2*time.Second)
	if err != nil {
		if err == context.DeadlineExceeded {
			logger.Warn("Timeout waiting for player transfer confirmation: ", player.ID().String())
			return errors.New("timeout waiting for player transfer confirmation")
		}
		return err
	}

	parts := strings.Split(msg.Payload, "|")
	if len(parts) == 4 && parts[0] == player.ID().String() && parts[1] == targetProxy {
		// server will be one of four things:
		// 0, means the proxy is found but given server is not available
		// 1, means the proxy is found and none server is specified
		// 2, means the proxy is found and the given server is available
		// 3, means the proxy is found and the given server is found but not responding
		server := parts[3]

		logger.Info(server)
		if server == "0" {
			logger.Warn("Specified server for player transfer was not found: ", player.ID().String())
			return errors.New("specified server was not found")
		}

		if server == "3" {
			logger.Warn("Specified server for player transfer was found, but is not responding: ", player.ID().String())
			return errors.New("specified server found but not responding")
		}

		if server == "2" {
			c := &cookie.Cookie{
				Key:     transferKey,
				Payload: []byte(targetServer),
			}

			err := cookie.Store(player, c)
			if err != nil {
				logger.Error("Error storing cookie to player: " + player.Username() + " - Error: " + err.Error())
				return errors.New("could not store cookie")
			}
		}

		address := parts[2]
		err := player.TransferToHost(address)
		if err != nil {
			return err
		}

		logger.Info("Player transfer successful: ", player.ID().String(), " to ", address)
	}

	return nil
}

func TransferPlayerToProxy(player proxy.Player, targetProxy string) error {
	return TransferPlayerToServerOnOtherProxy(player, targetProxy, "-")
}

func listenToTransfers() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database.StartListener(ctx, "proxy_transfer_request", func(msg *redis.Message) {
		parts := strings.Split(msg.Payload, "|")
		if len(parts) != 4 {
			logger.Error("Invalid transfer command format")
			return
		}

		targetProxy := parts[1]
		if targetProxy == proxy_name {

			playerID := parts[0]
			targetServer := parts[2]
			address := p.Config().Bind
			responseChannel := parts[3]

			server := "0"
			if targetServer == "-" {
				server = "1"
			} else {
				foundServer := p.Server(targetServer)
				if foundServer != nil {
					server = "2"
				}

				if !isBackendResponding(foundServer.ServerInfo().Addr().String()) {
					server = "3"
				}
			}

			err := database.Publish(context.Background(), responseChannel, playerID+"|"+targetProxy+"|"+address+"|"+server)
			if err != nil {
				logger.Warn("Error returning transfer. ", err)
			}
		}
	})
}
