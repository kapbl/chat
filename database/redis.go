package database

import (
	"github.com/redis/go-redis/v9"
)

var RedisDB *redis.Client

func InitRedis() {
	op := redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	rdb := redis.NewClient(&op)
	RedisDB = rdb
}
