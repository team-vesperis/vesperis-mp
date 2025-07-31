package database

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
)

type Database struct {
	r   *redis.Client
	p   *pgxpool.Pool
	l   *logger.Logger
	ctx context.Context
	lm  *listenManager
}

func Init(ctx context.Context, c *config.Config, l *logger.Logger) (*Database, error) {
	r, err := initRedis(ctx, l, c)
	if err != nil {
		return nil, err
	}

	p, err := initPostgres(ctx, l, c)
	if err != nil {
		return nil, err
	}

	lm := &listenManager{
		m: make(map[string]*redis.PubSub),
	}

	return &Database{
		r:   r,
		l:   l,
		lm:  lm,
		ctx: ctx,
		p:   p,
	}, nil
}

// handles all the listeners that are actively listening for messages
type listenManager struct {
	m  map[string]*redis.PubSub
	mu sync.Mutex
	wg sync.WaitGroup
}

const redisTTL = 15 * time.Minute

func (db *Database) GetData(key string) (any, error) {
	/*
		Plan:
		 1. Use Redis. If value is present, return.
		 2. If value is not present, use Postgres.
		 3. Update Redis if value wasn't present.
	*/

	var val any

	// 1.
	val, err := db.r.JSONGet(db.ctx, "data", "$."+key).Result()
	if err != nil && err != redis.Nil {
		db.l.Warn("redis get error", "key", key, "error", err)
	} else {
		return val, nil
	}

	// 2.
	query := "SELECT dataValue FROM data WHERE dataKey = $1"
	r, err := db.p.Query(db.ctx, query, key)
	if err != nil {
		db.l.Error("postgres query error", "key", key, "error", err)
		return "", err
	}
	defer r.Close()
	for r.Next() {
		scanErr := r.Scan(&val)
		if scanErr != nil {
			db.l.Error("postgres scan error", "key", key, "error", scanErr)
			return "", scanErr
		}

		if r.Err() != nil {
			db.l.Error("postgres rows error", "key", key, "error", r.Err())
			return "", r.Err()
		}
	}

	// 3.
	if val != "" {
		err := db.r.JSONSet(db.ctx, "data", "$."+key, val).Err()
		if err != nil {
			db.l.Warn("redis set error", "key", key, "error", err)
		}
	}

	return val, nil
}

func (db *Database) SetData(key string, val any) error {
	// redis
	err := db.r.JSONSet(db.ctx, "data", "$."+key, val).Err()
	if err != nil {
		db.l.Error("redis data set error", "key", key, "value", val, "error", err)
	}

	// postgres
	query := `
		INSERT INTO data 
		(dataKey, dataValue) VALUES ($1, $2) 
		ON CONFLICT (dataKey) DO UPDATE SET dataValue = $2
	`
	_, err = db.p.Exec(db.ctx, query, key, val)
	if err != nil {
		db.l.Error("postgres data upsert error", "key", key, "value", val, "error", err)
		return err
	}

	return nil
}

func (db *Database) SetPlayerData(playerId string, playerData map[string]any) error {
	// redis
	err := db.r.JSONSet(db.ctx, "player_data:"+playerId, "$", playerData).Err()
	if err != nil {
		db.l.Warn("redis set player data error", "playerId", playerId, "playerData", playerData, "error", err)
	}

	// postgres
	query := `
		INSERT INTO player_data
		(playerId, playerData) VALUES ($1, $2)
		ON CONFLICT (playerId) DO UPDATE SET playerData = $2
	`
	_, err = db.p.Exec(db.ctx, query, playerId, playerData)
	if err != nil {
		db.l.Error("postgres player data upsert error", "playerId", playerId, "playerData", playerData, "error", err)
		return err
	}

	return nil
}

func (db *Database) GetPlayerData(playerId string) (map[string]any, error) {
	playerData := make(map[string]any)

	// redis
	val, err := db.r.JSONGet(db.ctx, "player_data:"+playerId, "$").Result()
	if err != nil {
		db.l.Warn("redis get player data error", "playerId", playerId, "error", err)
	} else {
		if val != "" {
			err = json.Unmarshal([]byte(val), &playerData)
			if err != nil {
				db.l.Warn("json unmarshal player data error", "playerId", playerId, "error", err)
			} else {
				return playerData, nil
			}
		}
	}

	// postgres

	// update redis
	err = db.r.JSONSet(db.ctx, "player_data:"+playerId, "$", playerData).Err()
	if err != nil {
		db.l.Warn("redis set player data error", "playerId", playerId, "error", err)
	}

	return playerData, nil
}

func (db *Database) GetAllPlayerIds() ([]string, error) {
	var ids []string

	// redis

	// postgres

	return ids, nil
}

func (db *Database) SetPlayerDataField(playerId, field string, val any) error {
	// redis
	err := db.r.JSONSet(db.ctx, "player_data:"+playerId, "$."+field, val).Err()
	if err != nil {

	}

	// postgres

	return nil
}

func (db *Database) GetPlayerDataField(playerId, field string) (any, error) {
	var playerData any
	// redis
	val, err := db.r.JSONGet(db.ctx, "player_data:"+playerId, "$."+field).Result()
	if err != nil {

	} else {
		if val != "" {
			err = json.Unmarshal([]byte(val), &playerData)
			if err != nil {

			} else {
				return playerData, nil
			}
		}
	}

	// postgres

	// update redis
	err = db.r.JSONSet(db.ctx, "player_data:"+playerId, "$."+field, playerData).Err()
	if err != nil {

	}

	return nil, nil
}

func (db *Database) Publish(channel string, message any) error {
	cmd := db.r.Publish(db.ctx, channel, message)
	if cmd.Err() != nil {
		db.l.Error("redis publish error", "channel", channel, "message", message, "error", cmd.Err())
	}

	return nil
}

func (db *Database) Subscribe(channel string) *redis.PubSub {
	return db.r.Subscribe(db.ctx, channel)
}

func (db *Database) SubscribeWithTimeout(channel string, timeout time.Duration) (*redis.Message, error) {
	pubsub := db.r.Subscribe(db.ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

// Combination of Publish & Subscribe. Publish message in a channel, wait for a return message with a time limit.
func (db *Database) SendAndReturn(channel string, message any, timeout time.Duration) (*redis.Message, error) {
	pubsub := db.Subscribe(channel)
	defer pubsub.Close()

	err := db.Publish(channel, message)
	if err != nil {
		return nil, err
	}

	ch := pubsub.Channel()
	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

// Create a listener to listen for incoming calls. Is basically the same as the Subscribe function but it handles it for you. The listener can be stopped by using DeleteListener()
func (db *Database) CreateListener(channel string, handler func(msg *redis.Message)) {
	db.lm.mu.Lock()
	defer db.lm.mu.Unlock()

	if _, exists := db.lm.m[channel]; exists {
		db.l.Warn("database listener already existing", "channel", channel)
		return
	}

	pubsub := db.Subscribe(channel)
	db.lm.m[channel] = pubsub

	go func() {
		for {
			msg, err := pubsub.ReceiveMessage(db.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				db.l.Error("redis receiving message error", "channel", channel, "error", err)
				continue
			}
			db.lm.wg.Add(1)
			handler(msg)
			db.lm.wg.Done()
		}
	}()
}

func (db *Database) DeleteListener(channel string) error {
	db.lm.mu.Lock()
	defer db.lm.mu.Unlock()

	pubsub, exists := db.lm.m[channel]
	if !exists {
		return nil
	}

	delete(db.lm.m, channel)
	err := pubsub.Close()
	if err != nil {
		db.l.Error("redis closing pubsub error", "channel", channel, "error", err)
		return err
	}
	return nil
}

func (db *Database) DeleteAllListeners() error {
	db.lm.mu.Lock()
	defer db.lm.mu.Unlock()

	var firstErr error
	for channel := range db.lm.m {
		err := db.DeleteListener(channel)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	db.lm.m = make(map[string]*redis.PubSub)
	return firstErr
}

// Close the database. Closes the connection with Redis and PostgreSQL
func (db *Database) Close(ctx context.Context) {
	db.DeleteAllListeners()

	err := db.r.Close()
	if err != nil {
		db.l.Error("redis close error", "error", err)
	}

	ctx, canc := context.WithTimeout(ctx, 30*time.Second)
	defer canc()

	done := make(chan struct{})
	go func() {
		db.p.Close()
		close(done)
	}()

	select {
	case <-done:
		// closed successfully
	case <-ctx.Done():
		db.l.Error("postgres close timeout", "error", ctx.Err())
	}
}
