package database

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
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

func (db *Database) GetData(key string, dest any) error {
	val, err := db.r.HGet(db.ctx, DefaultDataType.String(), key).Result()
	if err == nil && val != "" {
		err := json.Unmarshal([]byte(val), dest)
		if err == nil {
			return nil
		}
	}

	var jsonData []byte
	query := `SELECT dataValue FROM data WHERE dataKey = $1`

	err = db.p.QueryRow(db.ctx, query, key).Scan(&jsonData)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrDataNotFound
		}
		db.l.Error("postgres data get error", "key", key, "error", err)
		return err
	}

	err = json.Unmarshal(jsonData, dest)
	if err != nil {
		db.l.Error("json data unmarshal error", "key", key, "error", err)
		return err
	}

	go func() {
		err = db.r.HSet(db.ctx, DefaultDataType.String(), key, jsonData).Err()
		if err != nil {
			db.l.Warn("redis data set error", "key", key, "error", err)
		} else {
			err = db.r.HExpire(db.ctx, DefaultDataType.String(), redisTTL, key).Err()
			if err != nil {
				db.l.Warn("redis data set expiration error", "key", key, "error", err)
			}
		}
	}()

	return nil
}

func (db *Database) SetData(key string, val any) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		db.l.Error("json data marshal error", "key", key, "error", err)
		return err
	}

	err = db.r.HSet(db.ctx, "data", key, jsonVal).Err()
	if err != nil {
		db.l.Warn("redis data set error", "key", key, "error", err)
	} else {
		err = db.r.HExpire(db.ctx, "data", redisTTL, key).Err()
		if err != nil {
			db.l.Warn("redis data set expiration error", "key", key, "error", err)
		}
	}

	query := `
		UPDATE data
		SET dataValue = $1
		WHERE dataKey = $2
	`
	_, err = db.p.Exec(db.ctx, query, jsonVal, key)
	if err != nil {
		db.l.Error("postgres data update error", "key", key, "error", err)
		return err
	}

	return nil
}

func (db *Database) setData(dt DataType, id uuid.UUID, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		db.l.Error("json "+dt.String()+" data marshal error", dt.String()+"Id", id, "error", err)
		return err
	}

	query := `INSERT INTO ` + dt.String() + `_data (` + dt.String() + `Id, ` + dt.String() + `Data) VALUES ($1, $2::jsonb) ON CONFLICT (` + dt.String() + `Id) DO UPDATE SET ` + dt.String() + `Data = $2::jsonb`
	_, err = db.p.Exec(db.ctx, query, id, jsonData)
	if err != nil {
		db.l.Error("postgres "+dt.String()+" data upsert error", dt.String()+"Id", id, "error", err)
		return err
	}

	return nil
}

func (db *Database) SetPlayerData(playerId uuid.UUID, data *data.PlayerData) error {
	return db.setData(PlayerDataType, playerId, data)
}

func (db *Database) SetPartyData(partyId uuid.UUID, data *data.PartyData) error {
	return db.setData(PartyDataType, partyId, data)
}

func (db *Database) SetProxyData(proxyId uuid.UUID, data *data.ProxyData) error {
	return db.setData(ProxyDataType, proxyId, data)
}

func (db *Database) SetBackendData(backendId uuid.UUID, data *data.BackendData) error {
	return db.setData(BackendDataType, backendId, data)
}

func (db *Database) getData(dt DataType, id uuid.UUID, dest any) error {
	query := `SELECT ` + dt.String() + `Data FROM ` + dt.String() + `_data WHERE ` + dt.String() + `Id = $1`
	err := db.p.QueryRow(db.ctx, query, id).Scan(dest)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrDataNotFound
		}
		db.l.Error("postgres "+dt.String()+" data read error", dt+"Id", id, "error", err)
		return err
	}

	return nil
}

func (db *Database) GetPlayerData(playerId uuid.UUID) (*data.PlayerData, error) {
	var data data.PlayerData
	err := db.getData(PlayerDataType, playerId, &data)
	if err != nil {
		return nil, err
	}

	if data.InitializeDefaults() {
		err := db.SetPlayerData(playerId, &data)
		if err != nil {
			db.l.Warn("failed to update player data with defaults", "playerId", playerId, "error", err)
		}
	}

	return &data, nil
}

func (db *Database) GetPartyData(partyId uuid.UUID) (*data.PartyData, error) {
	var data data.PartyData
	err := db.getData(PartyDataType, partyId, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (db *Database) GetProxyData(proxyId uuid.UUID) (*data.ProxyData, error) {
	var data data.ProxyData
	err := db.getData(ProxyDataType, proxyId, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (db *Database) GetBackendData(backendId uuid.UUID) (*data.BackendData, error) {
	var data data.BackendData
	err := db.getData(BackendDataType, backendId, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (db *Database) setDataField(dt DataType, id uuid.UUID, key, field string, val any) error {
	jsonVal, err := json.Marshal(val)
	if err != nil {
		db.l.Error("json "+dt.String()+" data marshal error", dt.String()+"Id", id, "field", field, "error", err)
		return err
	}

	go func() {
		err = db.r.HSet(db.ctx, key, field, jsonVal).Err()
		if err != nil {
			db.l.Warn("redis "+dt.String()+" data set error", dt.String()+"Id", id, "field", field, "error", err)
		} else {
			err = db.r.HExpire(db.ctx, key, redisTTL, field).Err()
			if err != nil {
				db.l.Warn("redis "+dt.String()+" data set expiration error", dt.String()+"Id", id, "field", field, "error", err)
			}
		}
	}()

	path := safeJsonPathForPostgres(field)
	query := "UPDATE " + dt.String() + "_data SET " + dt.String() + "Data = jsonb_set(" + dt.String() + "Data, $1, $2::jsonb, true)WHERE " + dt.String() + "Id = $3"
	_, err = db.p.Exec(db.ctx, query, path, jsonVal, id)
	if err != nil {
		db.l.Error("postgres "+dt.String()+" data update error", dt.String()+"Id", id, "field", field, "error", err)
		return err
	}

	return nil
}

func (db *Database) SetPlayerDataField(playerId uuid.UUID, field key.PlayerKey, val any) error {
	return db.setDataField(PlayerDataType, playerId, redisPlayerKeyTranslator(playerId), field.String(), val)
}

func (db *Database) SetPartyDataField(partyId uuid.UUID, field key.PartyKey, val any) error {
	return db.setDataField(PartyDataType, partyId, redisPartyKeyTranslator(partyId), field.String(), val)
}

func (db *Database) SetProxyDataField(proxyId uuid.UUID, field key.ProxyKey, val any) error {
	return db.setDataField(ProxyDataType, proxyId, redisProxyKeyTranslator(proxyId), field.String(), val)
}

func (db *Database) SetBackendDataField(backendId uuid.UUID, field key.BackendKey, val any) error {
	return db.setDataField(BackendDataType, backendId, redisBackendKeyTranslator(backendId), field.String(), val)
}

func (db *Database) getDataField(dt DataType, id uuid.UUID, key, field string, dest any) error {
	val, err := db.r.HGet(db.ctx, key, field).Result()
	if err == nil && val != "" {
		err := json.Unmarshal([]byte(val), dest)
		if err == nil {
			return nil
		}
	}

	var jsonData []byte
	query := "SELECT " + dt.String() + "Data #> $1 FROM " + dt.String() + "_data WHERE " + dt.String() + "Id = $2"

	path := safeJsonPathForPostgres(field)
	err = db.p.QueryRow(db.ctx, query, path, id).Scan(&jsonData)
	if err != nil {
		db.l.Error("postgres "+dt.String()+" data read error", dt+"Id", id, "error", err)
		return err
	}

	err = json.Unmarshal(jsonData, dest)
	if err != nil {
		db.l.Error("json "+dt.String()+" data unmarshall error", dt+"Id", id, "error", err)
		return err
	}

	go func() {
		err = db.r.HSet(db.ctx, key, field, jsonData).Err()
		if err != nil {
			db.l.Warn("redis "+dt.String()+" data set error", dt.String()+"Id", id, "field", field, "error", err)
		} else {
			err = db.r.HExpire(db.ctx, key, redisTTL, field).Err()
			if err != nil {
				db.l.Warn("redis "+dt.String()+" data set expiration error", dt.String()+"Id", id, "field", field, "error", err)
			}
		}
	}()

	return nil
}

func (db *Database) GetPlayerDataField(playerId uuid.UUID, field key.PlayerKey, dest any) error {
	return db.getDataField(PlayerDataType, playerId, redisPlayerKeyTranslator(playerId), field.String(), dest)
}

func (db *Database) GetPartyDataField(partyId uuid.UUID, field key.PartyKey, dest any) error {
	return db.getDataField(PartyDataType, partyId, redisPartyKeyTranslator(partyId), field.String(), dest)
}

func (db *Database) GetProxyDataField(proxyId uuid.UUID, field key.ProxyKey, dest any) error {
	return db.getDataField(ProxyDataType, proxyId, redisProxyKeyTranslator(proxyId), field.String(), dest)
}

func (db *Database) GetBackendDataField(backendId uuid.UUID, field key.BackendKey, dest any) error {
	return db.getDataField(BackendDataType, backendId, redisBackendKeyTranslator(backendId), field.String(), dest)
}

func (db *Database) DeletePartyData(partyId uuid.UUID) error {
	return db.deleteData(PartyDataType, partyId)

}
func (db *Database) DeleteProxyData(proxyId uuid.UUID) error {
	return db.deleteData(ProxyDataType, proxyId)

}

func (db *Database) DeleteBackendData(backendId uuid.UUID) error {
	return db.deleteData(BackendDataType, backendId)
}

func (db *Database) deleteData(dt DataType, id uuid.UUID) error {
	query := "DELETE FROM " + dt.String() + "_data WHERE " + dt.String() + "Id = $1"
	_, err := db.p.Exec(db.ctx, query, id)
	if err != nil {
		if err != pgx.ErrNoRows {
			return ErrDataNotFound
		}

		db.l.Error("postgres delete "+dt.String()+" data error", "error", err)
		return err
	}

	return nil
}

func (db *Database) GetAllPlayerIds() ([]uuid.UUID, error) {
	return db.getAllIds(PlayerDataType)
}

func (db *Database) GetAllPartyIds() ([]uuid.UUID, error) {
	return db.getAllIds(PartyDataType)
}

func (db *Database) GetAllProxyIds() ([]uuid.UUID, error) {
	return db.getAllIds(ProxyDataType)
}

func (db *Database) GetAllBackendsIds() ([]uuid.UUID, error) {
	return db.getAllIds(BackendDataType)
}

func (db *Database) getAllIds(dt DataType) ([]uuid.UUID, error) {
	query := "SELECT " + dt.String() + "Id FROM " + dt.String() + "_data"
	rows, err := db.p.Query(db.ctx, query)
	if err != nil {
		db.l.Error("postgres get all "+dt.String()+" ids error", "error", err)
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		err := rows.Scan(&id)
		if err != nil {
			db.l.Error("postgres scan "+dt.String()+" id error", "error", err)
			return nil, err
		}

		ids = append(ids, id)
	}
	if rows.Err() != nil {
		db.l.Error("postgres "+dt.String()+" rows error", "error", rows.Err())
		return nil, rows.Err()
	}

	return ids, nil
}

func (db *Database) AcquireLock(lockKey string, ttl time.Duration) (bool, error) {
	return db.r.SetNX(db.ctx, lockKey, "1", ttl).Result()
}

func (db *Database) ReleaseLock(lockKey string) error {
	return db.r.Del(db.ctx, lockKey).Err()
}

func (db *Database) Publish(channel string, message any) error {
	err := db.r.Publish(db.ctx, channel, message).Err()
	if err != nil {
		db.l.Error("redis publish error", "channel", channel, "message", message, "error", err)
	}

	return nil
}

func (db *Database) Subscribe(channel string) *redis.PubSub {
	db.l.Debug("database redis pubsub subscribing to channel", "channel", channel)
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
				db.l.Debug("database redis pubsub channel closed", "channel", channel)
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
func (db *Database) Close() error {
	err := db.DeleteAllListeners()
	if err != nil {
		db.l.Error("database deleting all listeners error", "error", err)
		return err
	}

	err = db.r.Close()
	if err != nil {
		db.l.Error("redis close error", "error", err)
		return err
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
		db.l.Info("database closed successfully")
		return nil
	case <-db.ctx.Done():
		db.l.Error("postgres close timeout", "error", db.ctx.Err())
		return db.ctx.Err()
	}
}
