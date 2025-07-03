package config

import (
	"os"

	"github.com/spf13/viper"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
)

const p = "./config/mp.yml"

type Config struct {
	v *viper.Viper // nil until created with the load function
	l *logger.Logger
}

func Init(l *logger.Logger) *Config {
	cfg := &Config{
		l: l,
	}

	cfg.load()
	return cfg
}

func (c *Config) load() error {
	c.v = viper.New()

	c.v.SetConfigName("mp")
	c.v.SetConfigType("yml")
	c.v.AddConfigPath("./config/")

	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		c.l.Warn("config file not found. creating default one...")
		err := c.createDefaultConfig()
		if err != nil {
			return err
		}
	}

	// test config
	err = c.v.ReadInConfig()
	if err != nil {
		return err
	}

	c.l.Info("loaded config")
	return nil
}

func (c *Config) GetViper() *viper.Viper {
	return c.v
}

func (c *Config) GetProxyId() string {
	return c.v.GetString("proxy_id")
}

func (c *Config) SetProxyId(id string) error {
	c.v.Set("proxy_id", id)
	return c.v.WriteConfig()
}

func (c *Config) SetBind(bind string) error {
	c.v.Set("config.bind", bind)
	return c.v.WriteConfig()
}

func (c *Config) GetRedisUrl() string {
	host := c.v.GetString("databases.redis.host")
	port := c.v.GetString("databases.redis.port")
	database := c.v.GetString("databases.redis.database")
	password := c.v.GetString("databases.redis.password")

	if password != "" {
		return "redis://:" + password + "@" + host + ":" + port + "/" + database
	}

	return "redis://" + host + ":" + port + "/" + database
}

func (c *Config) GetPostgresUrl() string {
	username := c.v.GetString("databases.postgres.username")
	password := c.v.GetString("databases.postgres.password")
	host := c.v.GetString("databases.postgres.host")
	port := c.v.GetString("databases.postgres.port")
	database := c.v.GetString("databases.postgres.database")

	return "postgres://" + username + ":" + password + "@" + host + ":" + port + "/" + database
}

func (c *Config) createDefaultConfig() error {
	defaultConfig := []byte(`
# The id that will be used to identify the proxy.
# If the id is already used it will override to a unique ID.
proxy_id: "proxy_123"

# The behavior of the gate proxy. By standard not needed, but it can be used to change behavior that is not changed by this program.
# config:


databases:
  redis:
    host: "localhost"
    port: 6379
    database: 0
    password: ""
  postgres:
    username: ""
    password: ""
    host: "localhost"
    port: 5432
    database: "vesperis_mp"
`)

	err := os.MkdirAll("./config", os.ModePerm)
	if err != nil {
		c.l.Error("config create directory error", "error", err)
		return err
	}

	err = os.WriteFile(p, defaultConfig, 0644)
	if err != nil {
		c.l.Error("config create default file error", "error", err)
		return err
	}

	c.l.Info("created default config")
	return nil
}
