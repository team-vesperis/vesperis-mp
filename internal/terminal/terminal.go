package terminal

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/team-vesperis/vesperis-mp/internal/mp/register"
	"github.com/team-vesperis/vesperis-mp/internal/playerdata"

	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "stop", Description: "Stop the proxy"},
		{Text: "register", Description: "Register a new server"},
		{Text: "unregister", Description: "Unregister a server"},
		{Text: "ban", Description: "Ban a player"},
		{Text: "unban", Description: "Unban a player"},
		{Text: "tempban", Description: "Temporarily ban a player"},
	}

	// Add subcommands
	text := d.TextBeforeCursor()
	if strings.HasPrefix(text, "register ") {
		s = []prompt.Suggest{
			{Text: "<server_name>", Description: "Name of the server"},
			{Text: "<host>", Description: "Host of the server"},
			{Text: "<port>", Description: "Port of the server"},
		}
	} else if strings.HasPrefix(text, "unregister ") {
		s = []prompt.Suggest{
			{Text: "<server_name>", Description: "Name of the server"},
		}
	} else if strings.HasPrefix(text, "ban ") {
		s = []prompt.Suggest{
			{Text: "<player_id>", Description: "ID of the player"},
			{Text: "<reason>", Description: "Reason for banning"},
		}
	} else if strings.HasPrefix(text, "unban ") {
		s = []prompt.Suggest{
			{Text: "<player_id>", Description: "ID of the player"},
		}
	} else if strings.HasPrefix(text, "tempban ") {
		s = []prompt.Suggest{
			{Text: "<player_id>", Description: "ID of the player"},
			{Text: "<reason>", Description: "Reason for banning"},
			{Text: "<duration_length>", Description: "Length of the ban duration"},
			{Text: "<duration_type>", Description: "Type of the ban duration (seconds, minutes, hours, days)"},
		}
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func HandleTerminalInput(p *proxy.Proxy, logger *zap.SugaredLogger) {
	time.Sleep(50 * time.Millisecond)

	for {
		cmd := prompt.Input("> ", completer)
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		if cmd == "stop" {
			p.Shutdown(&component.Text{
				Content: "The proxy has been manually stopped.",
			})
			return
		}

		if strings.HasPrefix(cmd, "register") {
			parts := strings.Split(cmd, " ")
			if len(parts) != 4 {
				logger.Info("Usage: register <server_name> <host> <port>")
				continue
			}
			server_name := parts[1]
			host := parts[2]
			port := parts[3]
			portInt, err := strconv.Atoi(port)
			if err != nil {
				logger.Error("Invalid port:", port)
				continue
			}

			register.RegisterServer(server_name, host, portInt)
		}

		if strings.HasPrefix(cmd, "unregister") {
			parts := strings.Split(cmd, " ")
			if len(parts) != 2 {
				logger.Info("Usage: unregister <server_name>")
				continue
			}
			server_name := parts[1]
			register.UnregisterServer(server_name)
		}

		if strings.HasPrefix(cmd, "ban") {
			parts := strings.Split(cmd, " ")
			if len(parts) != 3 {
				logger.Info("Usage: ban <player_id> <reason>")
				continue
			}
			playerId := parts[1]
			reason := parts[2]
			err := playerdata.BanPlayer(playerId, "", reason)
			if err != nil {
				logger.Error("Error banning player:", err)
			} else {
				logger.Info("Player banned:", playerId)
			}
		}

		if strings.HasPrefix(cmd, "unban") {
			parts := strings.Split(cmd, " ")
			if len(parts) != 2 {
				logger.Info("Usage: unban <player_id>")
				continue
			}
			playerId := parts[1]
			err := playerdata.UnBanPlayer(playerId)
			if err != nil {
				logger.Error("Error unbanning player:", err)
			} else {
				logger.Info("Player unbanned:", playerId)
			}
		}

		if strings.HasPrefix(cmd, "tempban") {
			parts := strings.Split(cmd, " ")
			if len(parts) != 5 {
				logger.Info("Usage: tempban <player_id> <reason> <duration_length> <duration_type>")
				continue
			}
			playerId := parts[1]
			reason := parts[2]
			durationLength, err := strconv.Atoi(parts[3])
			if err != nil {
				logger.Error("Invalid duration length:", parts[3])
				continue
			}
			durationType := parts[4]
			var duration time.Duration
			switch durationType {
			case "seconds":
				duration = time.Second
			case "minutes":
				duration = time.Minute
			case "hours":
				duration = time.Hour
			case "days":
				duration = time.Hour * 24
			default:
				logger.Error("Invalid duration type:", durationType)
				continue
			}
			err = playerdata.TempBanPlayer(playerId, "", reason, uint16(durationLength), duration)
			if err != nil {
				logger.Error("Error temp banning player:", err)
			} else {
				logger.Info("Player temp banned:", playerId)
			}
		}

		if strings.HasPrefix(cmd, "/") {
			log.Println("Command received:", cmd)
			dispatcher := p.Command().Dispatcher

			err := dispatcher.Do(context.Background(), strings.TrimPrefix(cmd, "/"))
			if err != nil {
				log.Println("Error executing command:", err)
			}
		}
	}
}
