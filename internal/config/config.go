package config

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var config *viper.Viper
var logger *zap.SugaredLogger

func LoadConfig(log *zap.SugaredLogger) *viper.Viper {
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
	return config
}

func createDefaultConfig() {
	defaultConfig := []byte(`
# The name that will be used to identify the proxy. 
# If the name is already used it will override to a unique ID.
proxy_name: "proxy_123"

databases:
  redis:
    host: "localhost"
    port: 6379
    database: 0
    username: ""
    password: ""
  postgresql:
	username: ""
	port: 5432
    host: "localhost"

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

func GetRedisUrl() string {
	host := config.GetString("databases.redis.host")
	port := config.GetString("databases.redis.port")
	database := config.GetString("databases.redis.database")
	password := config.GetString("databases.redis.password")

	if password != "" {
		return "redis://:" + password + "@" + host + ":" + port + "/" + database
	}

	return "redis://" + host + ":" + port + "/" + database
}

func GetPostgreSQL() string {
	username := config.GetString("databases.postgresql.username")
	password := config.GetString("databases.postgresql.password")
	host := config.GetString("databases.postgresql.host")
	port := config.GetString("databases.postgresql.port")
	database := config.GetString("databases.postgresql.database")

	return "postgres://" + username + ":" + password + "@" + host + ":" + port + "/" + database
}
