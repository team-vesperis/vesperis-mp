package share

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/mp/database"
)

func GetPlayerCount() int {
	value, err := database.GetRedisClient().Get(context.Background(), "player_count").Result()

	if err == redis.Nil {
		value = "0"
		updatePlayerCount(0)
	} else if err != nil {
		logger.Error("Error getting player count: ", err)
		return 0
	}
	playerCount, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		logger.Error("Error converting player count: ", err)
		return 0
	}
	return int(playerCount)
}

func AddPlayerToPlayerCount() {
	updatePlayerCount(1)
}

func RemovePlayerFromPlayerCount() {
	updatePlayerCount(-1)
}

func updatePlayerCount(delta int) {
	value, err := database.GetRedisClient().Get(context.Background(), "player_count").Result()
	if err == redis.Nil {
		logger.Warn("Player count key does not exist, initializing to 0")
		value = "0"
	} else if err != nil {
		logger.Error("Error getting player count: ", err)
		return
	}
	playerCount, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		logger.Error("Error converting player count: ", err)
		return
	}
	newCount := max(int(playerCount)+delta, 0)
	err = database.GetRedisClient().Set(context.Background(), "player_count", strconv.Itoa(newCount), 0).Err()
	if err != nil {
		logger.Error("Error setting player count: ", err)
	}
}
