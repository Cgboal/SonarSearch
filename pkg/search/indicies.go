package search

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
	"errors"
)

var redisClient *redis.Client
func init() {
	redisClient = newRedisClient()
}

func newRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		PoolSize: 64,
		DB:       0,
	})

	return rdb
}

func getPos(key string) (int64, error) {
	ctx := context.Background()
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return 0, errors.New("no results found")
	}

	valInt, _ := strconv.ParseInt(val, 10, 64)
	return valInt, nil
}
