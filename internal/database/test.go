package database

import (
	"time"

	"go.minekube.com/gate/pkg/util/uuid"
)

func (db *Database) TestDatabase() error {
	err := db.testData()
	if err != nil {
		return err
	}

	err = db.testPlayerData()
	if err != nil {
		return err
	}

	db.l.Info("Successful!")
	return nil
}

// don't use in production since there is an probability that the random id will override a real player's progress
func (db *Database) testPlayerData() error {
	id := uuid.New().String()

	now := time.Now()
	data, err1 := db.GetPlayerData(id)
	if err1 != nil {
		return err1
	}

	db.l.Info("returned data 1", "data", data, "duration", time.Since(now))

	data["test"] = "hi"
	data["online"] = "false"

	now = time.Now()
	err2 := db.SetPlayerData(id, data)
	if err2 != nil {
		return err2
	}
	db.l.Info("set data 2", "data", data, "duration", time.Since(now))

	now = time.Now()
	data2, err3 := db.GetPlayerData(id)
	if err3 != nil {
		return err3
	}

	db.l.Info("returned data 3", "data", data2, "duration", time.Since(now))

	now = time.Now()
	val, err4 := db.GetPlayerDataField(id, "not found!")
	if err4 != nil {
		return err4
	}

	db.l.Info("returned data field 4", "value", val, "duration", time.Since(now))

	return nil
}

func (db *Database) testData() error {
	key := uuid.New().Undashed()

	now := time.Now()
	val2, err1 := db.GetData(key)
	if err1 != nil {
		return err1
	}

	db.l.Info("returned value 1", "value", val2, "duration", time.Since(now))

	now = time.Now()
	val2, err2 := db.GetData(key)
	if err2 != nil {
		return err2
	}

	db.l.Info("returned value 2", "value", val2, "duration", time.Since(now))

	now = time.Now()
	err3 := db.SetData(key, "hi")
	if err3 != nil {
		return err3
	}

	db.l.Info("set value 3", "duration", time.Since(now))

	now = time.Now()
	val4, err4 := db.GetData(key)
	if err4 != nil {
		return err4
	}

	db.l.Info("returned value 4", "value", val4, "duration", time.Since(now))

	now = time.Now()
	val5, err5 := db.GetData(key)
	if err5 != nil {
		return err5
	}

	db.l.Info("returned value 5", "value", val5, "duration", time.Since(now))
	return nil
}
