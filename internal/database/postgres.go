package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"go.uber.org/zap"
)

func initializePostgres(ctx context.Context, l *zap.SugaredLogger) (*pgxpool.Pool, error) {
	p, err := pgxpool.New(ctx, config.GetPostgreSQL())
	if err != nil {
		l.Errorw("postgres connection error", "error", err)
		return nil, err
	}

	err = p.Ping(ctx)
	if err != nil {
		l.Errorw("postgres ping error", "error", err)
		return nil, err
	}

	err = createTables(ctx, p, l)
	if err != nil {
		l.Error("postgres creating table error")
		return nil, err
	}

	l.Info("Successfully initialized the Postgres Database.")
	return p, nil
}

func createTables(ctx context.Context, p *pgxpool.Pool, l *zap.SugaredLogger) error {
	playerDataTable := `
	CREATE TABLE IF NOT EXISTS player_data (
		playerId TEXT PRIMARY KEY,
		playerData JSONB NOT NULL
	);
	`

	_, err := p.Exec(ctx, playerDataTable)
	if err != nil {
		l.Errorw("postgres creating table error", "table", playerDataTable, "error", err)
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
		l.Errorw("postgres creating table error", "table", dataTable, "error", err)
		return err
	}

	return nil
}
