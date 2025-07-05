package database

import (
	"context"
	"database/sql"
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
	m   *sql.DB
	p   *pgxpool.Pool
	l   *logger.Logger
	ctx context.Context
	lm  *listenManager
}

func Init(ctx context.Context, c *config.Config, l *logger.Logger) (*Database, error) {
	r, err := initializeRedis(ctx, l, c)
	if err != nil {
		return nil, err
	}

	p, err := initializePostgres(ctx, l, c)
	if err != nil {
		return nil, err
	}

	lm := listenManager{
		m: make(map[string]*redis.PubSub),
	}

	return &Database{
		r:   r,
		l:   l,
		lm:  &lm,
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

func (d *Database) Get(key string) (any, error) {
	/*
		Plan:
		 1. Use Redis. If value is present, return.
		 2. If value is not present, use Postgres.
		 3. Update Redis if value wasn't present.
	*/

	// 1.
	cmd := d.r.Get(d.ctx, "data:"+key)
	if err := cmd.Err(); err == nil {
		val := cmd.Val()
		if val != "" {
			return val, nil
		}
	} else if err != redis.Nil {
		d.l.Error("redis get error", "key", key, "error", err)
	}

	// 2.
	var val any
	query := "SELECT dataValue FROM data WHERE dataKey = $1"
	r, err := d.p.Query(d.ctx, query, key)
	if err != nil {
		d.l.Error("postgres query error", "key", key, "error", err)
		return "", err
	}
	defer r.Close()
	for r.Next() {
		scanErr := r.Scan(&val)
		if scanErr != nil {
			d.l.Error("postgres scan error", "key", key, "error", scanErr)
			return "", scanErr
		}

		if r.Err() != nil {
			d.l.Error("postgres rows error", "key", key, "error", r.Err())
			return "", r.Err()
		}
	}

	// 3.
	if val != "" {
		err := d.r.Set(d.ctx, "data:"+key, val, redisTTL).Err()
		if err != nil {
			d.l.Warn("redis set error", "key", key, "error", err)
		}
	}

	return val, nil
}

func (d *Database) Set(key string, val any) error {
	// redis
	cmd := d.r.Set(d.ctx, "data:"+key, val, redisTTL)
	if cmd.Err() != nil {
		d.l.Error("redis set error", "key", key, "value", val, "error", cmd.Err())
	}

	// postgres
	query := "INSERT INTO data (dataKey, dataValue) VALUES ($1, $2) ON CONFLICT (dataKey) DO UPDATE SET dataValue = EXCLUDED.dataValue"
	_, err := d.p.Exec(d.ctx, query, key, val)
	if err != nil {
		d.l.Error("postgres upsert error", "key", key, "value", val, "error", err)
		return err
	}

	return nil
}

func (d *Database) SetPlayerData(playerId string, data map[string]any) error {
	// redis
	k := "player_data:" + playerId
	for key, val := range data {
		jsonVal, err := json.Marshal(val)
		if err == nil {
			_ = d.r.HSet(d.ctx, k, key, jsonVal).Err()
		}
	}

	d.r.Expire(d.ctx, k, redisTTL)

	// postgres
	jsonMap, err := json.Marshal(data)
	if err != nil {
		d.l.Error("marshal error", "data", data, "error", err)
		return err
	}

	query := `
		INSERT INTO player_data (playerId, playerData)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (playerId) DO UPDATE
		SET playerData = player_data.playerData || EXCLUDED.playerData
	`
	_, err = d.p.Exec(d.ctx, query, playerId, string(jsonMap))
	if err != nil {
		d.l.Error("postgres upsert error", "playerId", playerId, "error", err)
		return err
	}

	return nil
}

func (d *Database) GetPlayerData(playerId string) (map[string]any, error) {
	redisKey := "player_data:" + playerId
	playerData := make(map[string]any)

	// redis
	redisData, err := d.r.HGetAll(d.ctx, redisKey).Result()
	if err == nil && len(redisData) > 0 {
		for k, v := range redisData {
			var field any
			if json.Unmarshal([]byte(v), &field) == nil {
				playerData[k] = field
			}
		}
		return playerData, nil
	}

	// postgres
	var jsonData []byte
	query := "SELECT playerData FROM player_data WHERE playerId = $1"
	err = d.p.QueryRow(d.ctx, query, playerId).Scan(&jsonData)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[string]any{}, nil
		}
		d.l.Error("postgres read error", "playerId", playerId, "error", err)
		return nil, err
	}

	if err := json.Unmarshal(jsonData, &playerData); err != nil {
		d.l.Error("json unmarshal error", "playerId", playerId, "error", err)
		return nil, err
	}

	// update redis
	for k, v := range playerData {
		if j, err := json.Marshal(v); err == nil {
			_ = d.r.HSet(d.ctx, redisKey, k, j).Err()
		}
	}

	d.r.Expire(d.ctx, redisKey, redisTTL)

	return playerData, nil
}

func (d *Database) SetPlayerDataField(playerId, field string, val any) error {
	redisKey := "player_data:" + playerId

	// redis
	jsonVal, err := json.Marshal(val)
	if err == nil {
		_ = d.r.HSet(d.ctx, redisKey, field, jsonVal).Err()
	}

	d.r.Expire(d.ctx, redisKey, redisTTL)

	// postgres
	query := `
		INSERT INTO player_data (playerId, playerData)
		VALUES ($1, jsonb_build_object($2, to_jsonb($3)))
		ON CONFLICT (playerId) DO UPDATE
		SET playerData = jsonb_set(player_data.playerData, ARRAY[$2], to_jsonb($3), true)
	`
	_, err = d.p.Exec(d.ctx, query, playerId, field, val)
	if err != nil {
		d.l.Error("postgres update error", "playerId", playerId, "field", field, "error", err)
		return err
	}
	return nil
}

func (d *Database) GetPlayerDataField(playerId, field string) (any, error) {
	redisKey := "player_data:" + playerId

	// redis
	val, err := d.r.HGet(d.ctx, redisKey, field).Result()
	if err == nil && val != "" {
		var result any
		if json.Unmarshal([]byte(val), &result) == nil {
			return result, nil
		}
	}

	// postgres
	query := `
		SELECT playerData -> $2
		FROM player_data
		WHERE playerId = $1
	`
	var fieldData []byte
	err = d.p.QueryRow(d.ctx, query, playerId, field).Scan(&fieldData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		d.l.Error("postgres read error", "playerId", playerId, "field", field, "error", err)
		return nil, err
	}

	var result any
	if err := json.Unmarshal(fieldData, &result); err != nil {
		d.l.Error("json unmarshal error", "playerId", playerId, "field", field, "error", err)
		return nil, err
	}

	// update redis
	_ = d.r.HSet(d.ctx, redisKey, field, fieldData).Err()
	d.r.Expire(d.ctx, redisKey, redisTTL)

	return result, nil
}

func (d *Database) Publish(channel string, message any) error {
	cmd := d.r.Publish(d.ctx, channel, message)
	if cmd.Err() != nil {
		d.l.Error("redis publish error", "channel", channel, "message", message, "error", cmd.Err())
	}

	return nil
}

func (d *Database) Subscribe(channel string) *redis.PubSub {
	return d.r.Subscribe(d.ctx, channel)
}

func (d *Database) SubscribeWithTimeout(channel string, timeout time.Duration) (*redis.Message, error) {
	pubsub := d.r.Subscribe(d.ctx, channel)
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
func (d *Database) SendAndReturn(channel string, message any, timeout time.Duration) (*redis.Message, error) {
	pubsub := d.Subscribe(channel)
	defer pubsub.Close()

	err := d.Publish(channel, message)
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
func (d *Database) CreateListener(channel string, handler func(msg *redis.Message)) {
	d.lm.mu.Lock()
	defer d.lm.mu.Unlock()

	if _, exists := d.lm.m[channel]; exists {
		d.l.Warn("database listener already existing", "channel", channel)
		return
	}

	pubsub := d.Subscribe(channel)
	d.lm.m[channel] = pubsub

	go func() {
		for {
			msg, err := pubsub.ReceiveMessage(d.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				d.l.Error("redis receiving message error", "channel", channel, "error", err)
				continue
			}
			d.lm.wg.Add(1)
			handler(msg)
			d.lm.wg.Done()
		}
	}()
}

func (d *Database) DeleteListener(channel string) error {
	d.lm.mu.Lock()
	defer d.lm.mu.Unlock()

	pubsub, exists := d.lm.m[channel]
	if !exists {
		return nil
	}

	delete(d.lm.m, channel)
	err := pubsub.Close()
	if err != nil {
		d.l.Error("redis closing pubsub error", "channel", channel, "error", err)
		return err
	}
	return nil
}

func (d *Database) DeleteAllListeners() error {
	d.lm.mu.Lock()
	defer d.lm.mu.Unlock()

	var firstErr error
	for channel := range d.lm.m {
		err := d.DeleteListener(channel)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	d.lm.m = make(map[string]*redis.PubSub)
	return firstErr
}

// Close the database. Closes the connection with Redis and PostgreSQL
func (d *Database) Close() {
	d.DeleteAllListeners()

	err := d.r.Close()
	if err != nil {
		d.l.Error("redis close error", "error", err)
	}

	err = d.m.Close()
	if err != nil {
		d.l.Error("mysql close error", "error", err)
	}

	d.p.Close()
}
