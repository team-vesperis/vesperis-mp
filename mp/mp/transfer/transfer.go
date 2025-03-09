package transfer

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/mp/database"
	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p          *proxy.Proxy
	logger     *zap.SugaredLogger
	proxy_name string
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
			player.Disconnect(&component.Text{
				Content: "The proxy you were on has closed and there was no other proxy to connect to.",
				S: component.Style{
					Color: color.Red,
				},
			})
		} else {
			err = TransferPlayerToProxy(player, proxy)
			if err != nil {
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

// check if player has cookie specifying which server he needs.
func OnChooseInitialServer(event *proxy.PlayerChooseInitialServerEvent) {
	if len(p.Servers()) < 1 {
		event.Player().Disconnect(&component.Text{
			Content: "No available server. Please try again.",
			S: component.Style{
				Color: color.Red,
			},
		})
	} else {
		// payload, err := event.Player().RequestCookieWithResult(transferKey)
		// if err == nil {
		// 	server_name := string(payload)
		// 	server := p.Server(server_name)
		// 	if server != nil {
		// 		payload := []byte() // empty to reset cookie
		// 		err := player.StoreCookie(transferKey, payload)
		// 		if err != nil {
		// 			logger.Warn("Could not store cookie for player transfer for: ", player.ID().String())
		// 		}
		// 		event.SetInitialServer(server)
		// 	}
		// } else {
		event.SetInitialServer(p.Servers()[0])
		//}
	}
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
		// server will be one of three things:
		// 0, means the proxy is found but given server is not available
		// 1, means the proxy is found and none server is specified
		// 2, means the proxy is found and the given server is available
		server := parts[3]

		logger.Info(server)
		if server == "0" {
			logger.Warn("Specified server for player transfer was not found: ", player.ID().String())
			return errors.New("specified server was not found")
		}

		if server == "2" {
			// store cookie
			// payload := []byte(targetServer)
			// err := player.StoreCookie(transferKey, payload)
			// if err != nil {
			// 	logger.Warn("Could not store cookie for player transfer for: ", player.ID().String())
			// 	return errors.New("could not store cookie")
			// }
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
			}

			logger.Info(server)

			err := database.Publish(context.Background(), responseChannel, playerID+"|"+targetProxy+"|"+address+"|"+server)
			if err != nil {
				logger.Warn("Error returning transfer. ", err)
			}
		}
	})
}
