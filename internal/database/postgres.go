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

func createTable(ctx context.Context, p *pgxpool.Pool, dt DataType) error {
	table := "CREATE TABLE IF NOT EXISTS " + dt.String() + "_data (" + dt.String() + "Id UUID PRIMARY KEY," + dt.String() + "Data JSONB NOT NULL);"
	_, err := p.Exec(ctx, table)
	return err
}

func createTables(ctx context.Context, p *pgxpool.Pool, l *logger.Logger) error {
	err := createTable(ctx, p, PlayerDataType)
	if err != nil {
		l.Error("postgres creating player table error", "error", err)
		return err
	}

	err = createTable(ctx, p, PartyDataType)
	if err != nil {
		l.Error("postgres creating party table error", "error", err)
		return err
	}

	err = createTable(ctx, p, ProxyDataType)
	if err != nil {
		l.Error("postgres creating proxy table error", "error", err)
		return err
	}

	err = createTable(ctx, p, BackendDataType)
	if err != nil {
		l.Error("postgres creating backend table error", "error", err)
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
		l.Error("postgres creating data table error", "table", dataTable, "error", err)
		return err
	}

	return nil
}
