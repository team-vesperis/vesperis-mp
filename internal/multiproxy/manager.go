package multiproxy

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxyManager struct {
	multiProxyMap sync.Map

	// The context used in the mp.
	// Contains a cancel and logger.
	ctx context.Context

	// The logger used in the mp.
	l *logger.Logger

	// The database used in the multiproxy manager.
	// Contains a connection with Redis and Postgres.
	// Combines both in functions for fast and safe usage.
	db *database.Database

	// The config used in the mp.
	// Determines the database connection variables, proxy id, etc.
	c *config.Config

	// The multiplayer manager used in the mp.
	mpm *multiplayer.MultiPlayerManager
}

func InitManager(ctx context.Context) (*MultiProxyManager, error) {
	l, logErr := logger.Init()
	if logErr != nil {
		return &MultiProxyManager{}, logErr
	}

	c, cfErr := config.Init(l)
	if cfErr != nil {
		l.Error("config initialization error")
		return &MultiProxyManager{}, cfErr
	}

	db, dbErr := database.Init(ctx, c, l)
	if dbErr != nil {
		l.Error("database initialization error")
		return &MultiProxyManager{}, dbErr
	}

	mplayerm := multiplayer.InitManager(l, db)

	mproxym := &MultiProxyManager{
		multiProxyMap: sync.Map{},
		l:             l,
		c:             c,
		db:            db,
		ctx:           logr.NewContext(ctx, zapr.NewLogger(l.GetGateLogger())),
		mpm:           mplayerm,
	}

	return mproxym, nil
}

func (mpm *MultiProxyManager) GetMultiProxy(id uuid.UUID) (*MultiProxy, error) {
	val, ok := mpm.multiProxyMap.Load(id)
	if ok {
		mp, ok := val.(*MultiProxy)
		if ok {
			return mp, nil
		}

		mpm.multiProxyMap.Delete(id)
	}

	return nil, nil
}

func (mpm *MultiProxyManager) GetLogger() *logger.Logger {
	return mpm.l
}

func (mpm *MultiProxyManager) Close() {
	mpm.l.Info("Stopping...")
	mpm.db.Close()
	mpm.l.Close()
}
