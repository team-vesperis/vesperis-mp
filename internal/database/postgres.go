package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
)

func initPostgres(ctx context.Context, l *logger.Logger, c *config.Config) (*pgxpool.Pool, error) {
	now := time.Now()
	p, err := pgxpool.New(ctx, c.GetPostgresUrl())
	if err != nil {
		l.Error("postgres connection error", "error", err)
		return nil, err
	}

	err = p.Ping(ctx)
	if err != nil {
		l.Error("postgres ping error", "error", err)
		return nil, err
	}

	err = createTables(ctx, p, l)
	if err != nil {
		l.Error("postgres creating table error")
		return nil, err
	}

	l.Debug("initialized postgres", "duration", time.Since(now))
	return p, nil
}

func createTables(ctx context.Context, p *pgxpool.Pool, l *logger.Logger) error {
	playerDataTable := `
	CREATE TABLE IF NOT EXISTS player_data (
		playerId UUID PRIMARY KEY,
		playerData JSONB NOT NULL
	);
	`

	_, err := p.Exec(ctx, playerDataTable)
	if err != nil {
		l.Error("postgres creating table error", "table", playerDataTable, "error", err)
		return err
	}

	proxyDataTable := `
	CREATE TABLE IF NOT EXISTS proxy_data (
		proxyId UUID PRIMARY KEY,
		proxyData JSONB NOT NULL
	);
	`

	_, err = p.Exec(ctx, proxyDataTable)
	if err != nil {
		l.Error("postgres creating table error", "table", proxyDataTable, "error", err)
		return err
	}

	backendDataTable := `
	CREATE TABLE IF NOT EXISTS backend_data (
		backendId UUID PRIMARY KEY,
		backendData JSONB NOT NULL
	);
	`

	_, err = p.Exec(ctx, backendDataTable)
	if err != nil {
		l.Error("postgres creating table error", "table", backendDataTable, "error", err)
		return err
	}

	dataTable := `
	CREATE TABLE IF NOT EXISTS data (
		dataKey TEXT PRIMARY KEY,
		dataValue TEXT
	);
	`

	_, err = p.Exec(ctx, dataTable)
	if err != nil {
		l.Error("postgres creating table error", "table", dataTable, "error", err)
		return err
	}

	return nil
}
