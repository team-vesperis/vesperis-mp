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
	"go.minekube.com/gate/pkg/util/uuid"
)

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
	/*
		Plan:
		 1. Use Redis. If value is present, return.
		 2. If value is not present, use Postgres.
		 3. Update Redis if value wasn't present.
	*/

	var val any

	// 1.
	val, err := db.r.Get(db.ctx, "data:"+key).Result()
	if err != redis.Nil {
		if err != nil {
			db.l.Warn("redis data get error", "key", key, "error", err)
		} else {
			return val, nil
		}
	}

	// 2.
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

	// 3.
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

func redisKeyTranslator(playerId uuid.UUID, field string) string {
	return "player_data:" + playerId.String() + ":" + strings.ReplaceAll(field, ":", "_")
}

func safeJsonPath(field string) string {
	parts := strings.Split(field, ".")
	for i, part := range parts {
		if strings.ContainsAny(part, ` ."`) {
			parts[i] = `"` + strings.ReplaceAll(part, `"`, `\"`) + `"`
		}
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func buildNestedMap(flat map[string]any) map[string]any {
	nested := make(map[string]any)

	for flatKey, val := range flat {
		parts := strings.Split(flatKey, ".")
		curr := nested

		for i, part := range parts {
			if i == len(parts)-1 {
				curr[part] = val
			} else {
				if _, ok := curr[part]; !ok {
					curr[part] = make(map[string]any)
				}
				if next, ok := curr[part].(map[string]any); ok {
					curr = next
				} else {
					// Overwrite conflicting non-map value
					curr[part] = make(map[string]any)
					curr = curr[part].(map[string]any)
				}
			}
		}
	}

	return nested
}

func (db *Database) SetPlayerData(playerId uuid.UUID, playerData map[string]any) error {
	data := buildNestedMap(playerData)

	jsonData, err := json.Marshal(data)
	if err != nil {
		db.l.Error("json player data marshal error", "playerId", playerId, "error", err)
		return err
	}

	query := `
		INSERT INTO player_data (playerId, playerData)
		VALUES ($1, $2::jsonb)
		ON CONFLICT (playerId) DO 
		UPDATE SET playerData = $2::jsonb
	`
	_, err = db.p.Exec(db.ctx, query, playerId, jsonData)
	if err != nil {
		db.l.Error("postgres player data upsert error", "playerId", playerId, "error", err)
		return err
	}

	return nil
}

func (db *Database) GetPlayerData(playerId uuid.UUID) (map[string]any, error) {
	var jsonData []byte

	query := `
		SELECT playerData 
		FROM player_data 
		WHERE playerId = $1
	`
	err := db.p.QueryRow(db.ctx, query, playerId).Scan(&jsonData)
	if err != nil && err != pgx.ErrNoRows {
		db.l.Error("postgres player data read error", "playerId", playerId, "error", err)
		return nil, err
	}

	if jsonData == nil {
		return nil, ErrDataNotFound
	}

	var data map[string]any
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		db.l.Error("json player data unmarshal error", "playerId", playerId, "error", err)
		return nil, err
	}

	return data, nil
}

func (db *Database) SetPlayerDataField(playerId uuid.UUID, field string, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		db.l.Error("json player data marshal error", "playerId", playerId, "field", field, "error", err)
		return err
	}

	// Redis
	k := redisKeyTranslator(playerId, field)
	if redisErr := db.r.Set(db.ctx, k, data, redisTTL).Err(); redisErr != nil {
		db.l.Warn("redis player data set error", "playerId", playerId, "field", field, "error", redisErr)
	}

	// Postgres
	query := `
		UPDATE player_data
		SET playerData = jsonb_set(playerData, $1, $2::jsonb, true)
		WHERE playerId = $3
	`

	path := safeJsonPath(field)
	_, err = db.p.Exec(db.ctx, query, path, data, playerId)
	if err != nil {
		db.l.Error("postgres player data update error", "playerId", playerId, "field", field, "error", err)
		return err
	}

	return nil
}

func (db *Database) GetPlayerDataField(playerId uuid.UUID, field string) (any, error) {
	// Redis
	k := redisKeyTranslator(playerId, field)
	val, err := db.r.Get(db.ctx, k).Result()
	if err == nil && val != "" {
		var result any
		err := json.Unmarshal([]byte(val), &result)
		if err == nil {
			return result, nil
		}

		db.l.Warn("redis player data unmarshal error", "playerId", playerId, "field", field, "error", err)
	}

	// Postgres
	var fieldData []byte
	query := `
		SELECT playerData #> $1
		FROM player_data 
		WHERE playerId = $2
	`
	path := safeJsonPath(field)
	err = db.p.QueryRow(db.ctx, query, path, playerId).Scan(&fieldData)
	if err != nil && err != pgx.ErrNoRows {
		db.l.Error("postgres player data read error", "playerId", playerId, "field", field, "error", err)
		return nil, err
	}

	if fieldData == nil {
		return nil, ErrDataFieldNotFound
	}

	var result any
	err = json.Unmarshal(fieldData, &result)
	if err != nil {
		db.l.Error("json player data unmarshal error", "playerId", playerId, "field", field, "error", err)
		return nil, err
	}

	// cache redis
	err = db.r.Set(db.ctx, k, fieldData, redisTTL).Err()
	if err != nil {
		db.l.Warn("redis player data set error", "playerId", playerId, "field", field, "error", err)
	}

	return result, nil
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
		if err := rows.Scan(&id); err != nil {
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
