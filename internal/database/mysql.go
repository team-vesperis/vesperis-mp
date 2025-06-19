package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/team-vesperis/vesperis-mp/internal/config"
)

var database *sql.DB

func initializeMysql() {
	var err error
	database, err = sql.Open("mysql", config.GetMySQLUrl())
	if err != nil {
		logger.Panic("Error initializing MySQL Database. - ", err)
	}

	database.SetConnMaxLifetime(time.Minute * 3)
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(10)

	err = database.Ping()
	if err != nil {
		logger.Panic("Error pinging MySQL Database. - ", err)
	}

	createTables()
	logger.Info("Successfully initialized MySQL Database.")
}

func GetMySQLConnection(context context.Context) (*sql.Conn, error) {
	connection, err := database.Conn(context)
	if err != nil {
		logger.Panic("Error connecting with MySQL Database. - ", err)
		return connection, nil
	}

	return connection, nil
}

func createTables() {
	_, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS player_data (
			playerId VARCHAR(36) PRIMARY KEY,
			data JSON
		);
	`)

	if err != nil {
		logger.Panic("Error creating/loading player_data table. - ", err)
	}

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS banned_players (
    		playerId VARCHAR(36) PRIMARY KEY,
    		playerName VARCHAR(16),
    		reason TINYTEXT,
    		permanently BOOL,
    		ban_issued TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    		ban_expires TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)

	if err != nil {
		logger.Panic("Error creating/loading banned_players table. - ", err)
	}

	logger.Info("Successfully created/loaded MySQL table.")
}

func closeMySQL() {
	if database != nil {
		database.Close()
	}
}
