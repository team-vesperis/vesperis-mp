package multiplayer

import (
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
)

type MultiPlayerManager struct {
	multiPlayerMap sync.Map

	l  *logger.Logger
	db *database.Database
}

func InitMultiPlayerManager(l *logger.Logger, db *database.Database) *MultiPlayerManager {
	m := &MultiPlayerManager{
		multiPlayerMap: sync.Map{},
		l:              l,
		db:             db,
	}

	// fill map
	m.GetAllMultiPlayers()
	return m
}

func (m *MultiPlayerManager) GetAllMultiPlayers() []*MultiPlayer {
	var l []*MultiPlayer

	i, err := m.db.GetAllPlayerIds()
	if err != nil {
		return l
	}

	for _, id := range i {
		mp := m.GetMultiPlayer(id)
		l = append(l, mp)
	}

	return l
}

func (m *MultiPlayerManager) GetMultiPlayer(id string) *MultiPlayer {
	val, ok := m.multiPlayerMap.Load(id)
	if ok {
		mp, ok := val.(*MultiPlayer)
		if ok {
			return mp
		} else {
			m.multiPlayerMap.Delete(id)
		}
	}

	data, err := m.db.GetPlayerData(id)
	if err != nil || data == nil {
		return nil
	}

	mp := &MultiPlayer{
		id: id,
		m:  m,
	}

	if v, ok := data["p"].(string); ok {
		mp.p = v
	}
	if v, ok := data["b"].(string); ok {
		mp.b = v
	}
	if v, ok := data["name"].(string); ok {
		mp.name = v
	}
	if v, ok := data["online"].(bool); ok {
		mp.online = v
	}

	m.multiPlayerMap.Store(id, mp)
	return mp
}
