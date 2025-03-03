package playerdata

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/team-vesperis/vesperis-mp/mp/database"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func InitializePlayerData(log *zap.SugaredLogger) {
	logger = log
}

func setPlayerData(playerId string, playerData map[string]interface{}) {
	// redis
	redis := database.GetRedisClient()
	redisKey := "player_data:" + playerId

	for key, value := range playerData {
		jsonValue, _ := json.Marshal(value)

		err := redis.HSet(context.Background(), redisKey, key, jsonValue).Err()
		if err != nil {
			logger.Error("Error saving key and value to the Redis database. - ", err)
		}
	}

	// mysql
	mysql, err := database.GetMySQLConnection(context.Background())
	if err != nil {
		return
	}
	defer mysql.Close()

	jsonData, _ := json.Marshal(playerData)
	_, err = mysql.ExecContext(context.Background(), "REPLACE INTO player_data (playerId, data) VALUES (?, ?)", playerId, string(jsonData))
	if err != nil {
		logger.Error("Error saving player data to the MySQL database. - ", err)
	}
}

func setPlayerDataField(playerId, field string, value interface{}) {
	// redis
	redis := database.GetRedisClient()
	redisKey := "player_data:" + playerId

	jsonValue, _ := json.Marshal(value)
	err := redis.HSet(context.Background(), redisKey, field, jsonValue).Err()
	if err != nil {
		logger.Error("Error saving field to the Redis database. - ", err)
	}

	// mysql
	mysql, err := database.GetMySQLConnection(context.Background())
	if err != nil {
		return
	}
	defer mysql.Close()

	playerData := getPlayerData(playerId)
	playerData[field] = value
	jsonData, _ := json.Marshal(playerData)
	_, err = mysql.ExecContext(context.Background(), "REPLACE INTO player_data (playerId, data) VALUES (?, ?)", playerId, string(jsonData))
	if err != nil {
		logger.Error("Error saving player data field to the MySQL database. - ", err)
	}
}

func getPlayerDataField(playerId, field string) interface{} {
	// redis
	redis := database.GetRedisClient()
	redisKey := "player_data:" + playerId

	value, err := redis.HGet(context.Background(), redisKey, field).Result()
	if err == nil {
		var result interface{}
		json.Unmarshal([]byte(value), &result)
		return result
	}

	// mysql
	mysql, err := database.GetMySQLConnection(context.Background())
	if err != nil {
		return nil
	}
	defer mysql.Close()

	var jsonData string
	err = mysql.QueryRowContext(context.Background(), "SELECT data FROM player_data WHERE playerId = ?", playerId).Scan(&jsonData)
	if err != nil {

		// no data found -> create new data
		if err == sql.ErrNoRows {
			emptyData := make(map[string]interface{})
			setPlayerData(playerId, emptyData)
			return nil
		}

		logger.Error("Error getting player data field from the MySQL database. - ", err)
		return nil
	}

	// update redis
	var playerData map[string]interface{}
	json.Unmarshal([]byte(jsonData), &playerData)
	for key, value := range playerData {
		jsonValue, _ := json.Marshal(value)
		redis.HSet(context.Background(), redisKey, key, jsonValue)
	}

	return playerData[field]
}

func getPlayerData(playerId string) map[string]interface{} {
	// redis
	redis := database.GetRedisClient()
	redisKey := "player_data:" + playerId

	redisData, err := redis.HGetAll(context.Background(), redisKey).Result()
	if err == nil && len(redisData) > 0 {
		playerData := make(map[string]interface{})
		for key, value := range redisData {
			var fieldValue interface{}
			json.Unmarshal([]byte(value), &fieldValue)
			playerData[key] = fieldValue
		}
		return playerData
	}

	// mysql
	mysql, err := database.GetMySQLConnection(context.Background())
	if err != nil {
		return nil
	}
	defer mysql.Close()

	var jsonData string
	err = mysql.QueryRowContext(context.Background(), "SELECT data FROM player_data WHERE playerId = ?", playerId).Scan(&jsonData)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("No player data found in MySQL for player: ", playerId)
			// Create new empty data
			emptyData := make(map[string]interface{})
			setPlayerData(playerId, emptyData)
			return emptyData
		}
		logger.Error("Error getting player data from the MySQL database. - ", err)
		return nil
	}

	// update redis
	var playerData map[string]interface{}
	json.Unmarshal([]byte(jsonData), &playerData)
	for key, value := range playerData {
		jsonValue, _ := json.Marshal(value)
		redis.HSet(context.Background(), redisKey, key, jsonValue)
	}

	return playerData
}
