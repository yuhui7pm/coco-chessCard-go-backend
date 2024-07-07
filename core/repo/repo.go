package repo

import (
	"common/database"
)

type Manager struct {
	Mongo *database.MongoManager
	Redis *database.RedisManager
}

func New() *Manager {
	return &Manager{
		Mongo: database.NewMongo(),
		Redis: database.NewRedis(),
	}
}
