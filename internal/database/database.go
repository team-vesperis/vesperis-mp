package database

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/util/uuid"
)

// The Database uses Postgres as main storage. Whenever possible, redis will help, to make searches faster. Redis is used for communication between proxies.
type Database struct {
	r   *redis.Client
	p   *pgxpool.Pool
	l   *logger.Logger
	ctx context.Context
	lm  *listenManager
}

var (
	ErrDataFieldNotFound  = errors.New("data field not found")
	ErrDataNotFound       = errors.New("data not found")
	ErrIncorrectValueType = errors.New("incorrect value type returned from database")
)

func Init(ctx context.Context, c *config.Config, l *logger.Logger) (*Database, error) {
	now := time.Now()

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

	db := &Database{
		r:   r,
		l:   l,
		lm:  lm,
		ctx: ctx,
		p:   p,
	}

	// err = db.Test()
	// if err != nil {
	// 	return db, err
	// }

	db.l.Info("initialized database", "duration", time.Since(now))
	return db, nil
}

// handles all the listeners that are actively listening for messages
type listenManager struct {
	m  map[string]*redis.PubSub
	mu sync.Mutex
}

const redisTTL = 15 * time.Minute

func (db *Database) GetData(key string) (any, error) {
	var val any

	val, err := db.r.Get(db.ctx, "data:"+key).Result()
	if err != redis.Nil {
		if err != nil {
			db.l.Warn("redis data get error", "key", key, "error", err)
		} else {
			return val, nil
		}
	}

	query := "SELECT dataValue FROM data WHERE dataKey = $1"
	r, err := db.p.Query(db.ctx, query, key)
	if err != nil {
		db.l.Error("postgres data query error", "key", key, "error", err)
		return "", err
	}
	defer r.Close()
	for r.Next() {
		scanErr := r.Scan(&val)
		if scanErr != nil {
			db.l.Error("postgres data scan error", "key", key, "error", scanErr)
			return "", scanErr
		}

		if r.Err() != nil {
			db.l.Error("postgres data rows error", "key", key, "error", r.Err())
			return "", r.Err()
		}
	}

	if val == nil || val == "" {
		return "", ErrDataNotFound
	}

	err = db.r.Set(db.ctx, "data:"+key, val, redisTTL).Err()
	if err != nil {
		db.l.Warn("redis data set error", "key", key, "error", err)
	}

	return val, nil
}

func (db *Database) SetData(key string, val any) error {
	// redis
	err := db.r.Set(db.ctx, "data:"+key, val, redisTTL).Err()
	if err != nil {
		db.l.Warn("redis data set error", "key", key, "value", val, "error", err)
	}

	// postgres
	query := `
		INSERT INTO data (dataKey, dataValue) 
		VALUES ($1, $2) 
		ON CONFLICT (dataKey) DO 
		UPDATE SET dataValue = $2
	`
	_, err = db.p.Exec(db.ctx, query, key, val)
	if err != nil {
		db.l.Error("postgres data upsert error", "key", key, "value", val, "error", err)
		return err
	}

	return nil
}

func redisPlayerKeyTranslator(playerId uuid.UUID) string {
	return "player_data:" + playerId.String()
}

func redisProxyKeyTranslator(proxyId uuid.UUID) string {
	return "proxy_data:" + proxyId.String()
}

func safeJsonPathForPostgres(field string) string {
	parts := strings.Split(field, ".")
	for i, part := range parts {
		if strings.ContainsAny(part, ` ."`) {
			parts[i] = `"` + strings.ReplaceAll(part, `"`, `\"`) + `"`
		}
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (db *Database) SetPlayerData(playerId uuid.UUID, data *util.PlayerData) error {
	query := `
		INSERT INTO player_data (playerId, playerData)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (playerId) DO UPDATE
		SET playerData = $2::jsonb
	`
	_, err := db.p.Exec(db.ctx, query, playerId, data)
	if err != nil {
		db.l.Error("postgres player data upsert error", "playerId", playerId, "error", err)
		return err
	}

	return nil
}

func (db *Database) SetProxyData(proxyId uuid.UUID, data *util.ProxyData) error {
	query := `
		INSERT INTO proxy_data (proxyId, proxyData)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (proxyId) DO UPDATE
		SET proxyData = $2::jsonb
	`
	_, err := db.p.Exec(db.ctx, query, proxyId, data)
	if err != nil {
		db.l.Error("postgres proxy data upsert error", "proxyId", proxyId, "error", err)
		return err
	}

	return nil
}

func (db *Database) GetPlayerData(playerId uuid.UUID) (*util.PlayerData, error) {
	var data util.PlayerData
	query := `SELECT playerData FROM player_data WHERE playerId = $1`
	err := db.p.QueryRow(db.ctx, query, playerId).Scan(&data)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrDataNotFound
		}
		db.l.Error("postgres player data read error", "playerId", playerId, "error", err)
		return nil, err
	}

	return &data, nil
}

func (db *Database) GetProxyData(proxyId uuid.UUID) (*util.ProxyData, error) {
	var data util.ProxyData
	query := `SELECT proxyData FROM proxy_data WHERE proxyId = $1`
	err := db.p.QueryRow(db.ctx, query, proxyId).Scan(&data)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrDataNotFound
		}
		db.l.Error("postgres proxy data read error", "playerId", proxyId, "error", err)
		return nil, err
	}

	return &data, nil
}

func (db *Database) SetPlayerDataField(playerId uuid.UUID, field util.PlayerKey, val any) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		db.l.Error("json player data marshal error", "playerId", playerId, "field", field, "error", err)
		return err
	}

	key := redisPlayerKeyTranslator(playerId)
	err = db.r.HSet(db.ctx, key, field.String(), jsonVal).Err()
	if err != nil {
		db.l.Warn("redis player data set error", "playerId", playerId, "field", field, "error", err)
	} else {
		err = db.r.HExpire(db.ctx, key, redisTTL, field.String()).Err()
		if err != nil {
			db.l.Warn("redis player data set expiration error", "playerId", playerId, "field", field, "error", err)
		}
	}

	path := safeJsonPathForPostgres(field.String())
	query := `
		UPDATE player_data
		SET playerData = jsonb_set(playerData, $1, $2::jsonb, true)
		WHERE playerId = $3
	`
	_, err = db.p.Exec(db.ctx, query, path, jsonVal, playerId)
	if err != nil {
		db.l.Error("postgres player data update error", "playerId", playerId, "field", field, "error", err)
		return err
	}

	return nil
}

func (db *Database) SetProxyDataField(proxyId uuid.UUID, field util.ProxyKey, val any) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		db.l.Error("json proxy data marshal error", "proxyId", proxyId, "field", field, "error", err)
		return err
	}

	key := redisProxyKeyTranslator(proxyId)
	err = db.r.HSet(db.ctx, key, field.String(), jsonVal).Err()
	if err != nil {
		db.l.Warn("redis proxy data set error", "proxyId", proxyId, "field", field, "error", err)
	} else {
		err = db.r.HExpire(db.ctx, key, redisTTL, field.String()).Err()
		if err != nil {
			db.l.Warn("redis proxy data set expiration error", "proxyId", proxyId, "field", field, "error", err)
		}
	}

	path := safeJsonPathForPostgres(field.String())
	query := `
		UPDATE proxy_data
		SET proxyData = jsonb_set(proxyData, $1, $2::jsonb, true)
		WHERE proxyId = $3
	`
	_, err = db.p.Exec(db.ctx, query, path, jsonVal, proxyId)
	if err != nil {
		db.l.Error("postgres proxy data update error", "proxyId", proxyId, "field", field, "error", err)
		return err
	}

	return nil
}

func (db *Database) GetPlayerDataField(playerId uuid.UUID, field util.PlayerKey, dest any) error {
	key := redisPlayerKeyTranslator(playerId)

	val, err := db.r.HGet(db.ctx, key, field.String()).Result()
	if err == nil && val != "" {
		err := json.Unmarshal([]byte(val), dest)
		if err == nil {
			return nil
		}
	}

	var jsonData []byte
	query := `SELECT playerData #> $1 FROM player_data WHERE playerId = $2`

	path := safeJsonPathForPostgres(field.String())
	err = db.p.QueryRow(db.ctx, query, path, playerId).Scan(&jsonData)
	if err != nil {
		db.l.Error("")
		return err
	}

	err = json.Unmarshal(jsonData, dest)
	if err != nil {
		db.l.Error("")
		return err
	}

	err = db.r.HSet(db.ctx, key, field, jsonData).Err()
	if err != nil {
		db.l.Warn("")
	} else {
		err = db.r.Expire(db.ctx, key, redisTTL).Err()
		if err != nil {
			db.l.Warn("")
		}
	}

	return nil
}

func (db *Database) GetProxyDataField(proxyId uuid.UUID, field util.ProxyKey, dest any) error {
	key := redisProxyKeyTranslator(proxyId)

	val, err := db.r.HGet(db.ctx, key, field.String()).Result()
	if err == nil && val != "" {
		err := json.Unmarshal([]byte(val), dest)
		if err == nil {
			return nil
		}
	}

	var jsonData []byte
	query := `SELECT proxyData #> $1 FROM proxy_data WHERE proxyId = $2`

	path := safeJsonPathForPostgres(field.String())
	err = db.p.QueryRow(db.ctx, query, path, proxyId).Scan(&jsonData)
	if err != nil {
		db.l.Error("")
		return err
	}

	err = json.Unmarshal(jsonData, dest)
	if err != nil {
		db.l.Error("")
		return err
	}

	err = db.r.HSet(db.ctx, key, field, jsonData).Err()
	if err != nil {
		db.l.Warn("")
	} else {
		err = db.r.Expire(db.ctx, key, redisTTL).Err()
		if err != nil {
			db.l.Warn("")
		}
	}

	return nil
}

func (db *Database) GetAllPlayerIds() ([]uuid.UUID, error) {
	query := `SELECT playerId FROM player_data`
	rows, err := db.p.Query(db.ctx, query)
	if err != nil {
		db.l.Error("postgres get all player ids error", "error", err)
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		err := rows.Scan(&id)
		if err != nil {
			db.l.Error("postgres scan player id error", "error", err)
			return nil, err
		}

		ids = append(ids, id)
	}
	if rows.Err() != nil {
		db.l.Error("postgres rows error", "error", rows.Err())
		return nil, rows.Err()
	}

	return ids, nil
}

func (db *Database) GetAllProxyIds() ([]uuid.UUID, error) {
	query := `SELECT proxyId FROM proxy_data`
	rows, err := db.p.Query(db.ctx, query)
	if err != nil {
		db.l.Error("postgres get all player ids error", "error", err)
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		err := rows.Scan(&id)
		if err != nil {
			db.l.Error("postgres scan player id error", "error", err)
			return nil, err
		}

		ids = append(ids, id)
	}
	if rows.Err() != nil {
		db.l.Error("postgres rows error", "error", rows.Err())
		return nil, rows.Err()
	}

	return ids, nil
}

func (db *Database) Publish(channel string, message any) error {
	err := db.r.Publish(db.ctx, channel, message).Err()
	if err != nil {
		db.l.Error("redis publish error", "channel", channel, "message", message, "error", err)
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
func (db *Database) SendAndReturn(publishChannel, subscribeChannel string, message any, timeout time.Duration) (*redis.Message, error) {
	pubsub := db.Subscribe(subscribeChannel)
	defer pubsub.Close()

	err := db.Publish(publishChannel, message)
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
			msg, ok := <-pubsub.Channel()
			if !ok {
				return
			}

			handler(msg)
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
	listeners := db.lm.m
	db.lm.mu.Unlock()

	var firstErr error
	for channel := range listeners {
		err := db.DeleteListener(channel)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	db.lm.mu.Lock()
	db.lm.m = make(map[string]*redis.PubSub)
	db.lm.mu.Unlock()
	return firstErr
}

// Close the database. Closes the connection with Redis and PostgreSQL
func (db *Database) Close() {
	err := db.DeleteAllListeners()
	if err != nil {
		db.l.Error("database deleting all listeners error", "error", err)
	}

	err = db.r.Close()
	if err != nil {
		db.l.Error("redis close error", "error", err)
	}

	var canc context.CancelFunc
	db.ctx, canc = context.WithTimeout(db.ctx, 30*time.Second)
	defer canc()

	done := make(chan struct{})
	go func() {
		db.p.Close()
		close(done)
	}()

	select {
	case <-done:
		// closed successfully
		return
	case <-db.ctx.Done():
		db.l.Error("postgres close timeout", "error", db.ctx.Err())
		return
	}
}
