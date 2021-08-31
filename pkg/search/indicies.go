package search

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
	"fmt"
)

func newRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return rdb
}

func getPos(key string) (int64, error) {
	client := newRedisClient()
	defer client.Close()
	ctx := context.Background()
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	fmt.Println(val)

	valInt, _ := strconv.ParseInt(val, 10, 64)
	return valInt, nil
}
