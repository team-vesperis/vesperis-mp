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
		if err1 == ErrDataNotFound {
			// create default
			db.SetPlayerData(id, map[string]any{
				"not anymore": true,
			})
		} else {
			return err1

		}
	}

	db.l.Info("returned data 1", "data", data, "duration", time.Since(now))

	data["test"] = "hi"
	data["online"] = false

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
	val, err4 := db.GetPlayerDataField(id, "online")
	if err4 != nil {
		return err4
	}

	if val == false {
		db.l.Info("correctly turned value into bool")
	}

	db.l.Info("returned data field 4", "value", val, "duration", time.Since(now))

	now = time.Now()
	val2, err5 := db.GetPlayerDataField(id, "not_found")
	if err5 != nil {
		// value not found -> set default
		if err5 == ErrDataFieldNotFound {
			data := []string{
				"hello",
				"more strings",
				"this saves correctly",
			}

			db.SetPlayerDataField(id, "not_found", data)
		} else {
			db.l.Info("error 5", "error", err5)
			return err5 // unknown error. handle accordingly
		}
	}

	db.l.Info("returned data field 5", "value", val2, "duration", time.Since(now))

	val, err6 := db.GetPlayerDataField(id, "not_found")
	if err6 != nil {
		return err6
	}

	list, ok := val.([]any)
	if ok {

		for _, a := range list {
			s, ok := a.(string)
			if ok {
				db.l.Info(s)
			}
		}
	}

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
