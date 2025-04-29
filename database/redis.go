package database

import (
	"context"
	"log"

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
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	log.Println(pong)
	RedisDB = rdb
}
