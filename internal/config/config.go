package config

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var config *viper.Viper
var logger *zap.SugaredLogger

func LoadConfig(log *zap.SugaredLogger) {
	logger = log
	config = viper.New()

	config.SetConfigName("mp")
	config.SetConfigType("yml")
	config.AddConfigPath("./config")

	if _, err := os.Stat("./config/mp.yml"); os.IsNotExist(err) {
		logger.Warn("Config file not found, creating default...")
		createDefaultConfig()
	}

	if err := config.ReadInConfig(); err != nil {
		logger.Fatal("Error reading config. - ", err)
	}

	logger.Info("Successfully loaded config.")
}

func createDefaultConfig() {
	defaultConfig := []byte(`
# The name that will be used to identify the proxy in the redis database. 
# If the name is already used it will override to a unique ID.
proxy_name: "proxy_123"

databases:
  mysql:
    username: ""
    password: ""
    host: "localhost"
    port: 3306
    database: "vesperis_mp"
  redis:
    host: "localhost"
    port: 6379
    database: 0
    username: ""
    password: ""
`)

	if err := os.MkdirAll("./config", os.ModePerm); err != nil {
		logger.Panic("Failed to create config directory. - " + err.Error())
	}

	if err := os.WriteFile("./config/mp.yml", defaultConfig, 0644); err != nil {
		logger.Panic("Failed to create default config. - " + err.Error())
	}

	logger.Info("Successfully created config.")
}

func GetProxyName() string {
	return config.GetString("proxy_name")
}

func SetProxyName(newName string) {
	config.Set("proxy_name", newName)
}

func GetMySQLUrl() string {
	return config.GetString("databases.mysql.username") +
		":" +
		config.GetString("databases.mysql.password") +
		"@(" +
		config.GetString("databases.mysql.host") +
		":" +
		config.GetString("databases.mysql.port") +
		")/" +
		config.GetString("databases.mysql.database") +
		"?parseTime=true"
}

func GetRedisUrl() string {
	host := config.GetString("databases.redis.host")
	port := config.GetString("databases.redis.port")
	database := config.GetString("databases.redis.database")
	username := config.GetString("databases.redis.username")
	password := config.GetString("databases.redis.password")

	if username != "" && password != "" {
		return "redis://" + username + ":" + password + "@" + host + ":" + port + "/" + database
	}

	return "redis://" + host + ":" + port + "/" + database
}
