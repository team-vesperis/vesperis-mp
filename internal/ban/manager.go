package ban

import (
	"context"
	"fmt"
	"time"

	redislock "github.com/jefferyjob/go-redislock"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/playerdata"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	logger     *zap.SugaredLogger
	quit       chan struct{} = make(chan struct{})
	banChecker               = 5 * time.Minute
	ctx                      = context.Background()
	lock       redislock.RedisLockInter
)

func InitializeBanManager(log *zap.SugaredLogger) {
	logger = log
	lock = redislock.New(ctx, database.GetRedisClient(), "ban_key")

	logger.Info("Initialized Ban Manager.")

	go func() {
		for {
			err := lock.Lock()
			// another proxy is already running the checker
			if err != nil {
				logger.Error("Error acquiring lock for ban checker: ", err)
				time.Sleep(banChecker)
				continue
			}

			ticker := time.NewTicker(banChecker)
			for {
				select {
				case <-ticker.C:
					logger.Info("Checking temp bans for expired ones...")
					playerdata.CheckTempBans()
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}
	}()
}

func CloseBanManager() {
	lock.UnLock()
	close(quit)
}

func IsPlayerBanned(player proxy.Player) bool {
	return playerdata.IsPlayerBanned(player.ID().String())
}

func IsPlayerPermanentlyBanned(player proxy.Player) bool {
	return playerdata.IsPlayerPermanentlyBanned(player.ID().String())
}

func GetBanReason(player proxy.Player) string {
	reason := playerdata.GetBanReason(player.ID().String())
	if reason == "" {
		return "No reason provided."
	}

	return reason
}

func BanPlayer(player proxy.Player, reason string) {
	player.Disconnect(&component.Text{
		Content: "You are permanently banned from VesperisMC.",
		S: component.Style{
			Color: color.Red,
		},
		Extra: []component.Component{
			&component.Text{
				Content: "\n\nReason: " + reason,
				S: component.Style{
					Color: color.Gray,
				},
			},
		},
	})

	logger.Info("Player " + player.Username() + " - " + player.ID().String() + " has been banned. Reason: " + reason)
	playerdata.BanPlayer(player.ID().String(), player.Username(), reason)
}

func TempBanPlayer(player proxy.Player, reason string, durationLength uint16, durationType time.Duration) {
	duration := time.Duration(durationLength) * durationType
	hours := int(duration.Hours())
	days := hours / 24
	hours = hours % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	time := fmt.Sprintf("%d days, %d hours, %d minutes and %d seconds", days, hours, minutes, seconds)

	player.Disconnect(&component.Text{
		Content: "You are temporarily banned from VesperisMC",
		S: component.Style{
			Color: color.Red,
		},
		Extra: []component.Component{
			&component.Text{
				Content: "\n\nReason: " + reason,
				S: component.Style{
					Color: color.Gray,
				},
			},
			&component.Text{
				Content: "\n\nYou are banned for " + time,
				S: component.Style{
					Color: color.Aqua,
				},
			},
		},
	})

	logger.Info("Player " + player.Username() + " - " + player.ID().String() + " has been temporarily banned. Reason: " + reason + " Time: " + time)
	playerdata.TempBanPlayer(player.ID().String(), player.Username(), reason, durationLength, durationType)
}

func UnBanPlayer(playerId string) {
	logger.Info("Player with ID: " + playerId + " has been manually unbanned.")
	playerdata.UnBanPlayer(playerId)
}

func GetBanExpiration(player proxy.Player) time.Time {
	return playerdata.GetBanExpiration(player.ID().String())
}

func GetBannedPlayerNameList() []string {
	return playerdata.GetBannedPlayerNameList()
}

func GetBannedPlayerIdByName(playerName string) string {
	return playerdata.GetBannedPlayerIdByName(playerName)
}
