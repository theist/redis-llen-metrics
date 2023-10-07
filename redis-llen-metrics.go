package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()

func ListRedisLists() []string {
	res := []string{}

	for _, key := range rdb.Keys(ctx, "*").Val() {
		redisType := rdb.Type(ctx, key).Val()
		if redisType == "list" {
			res = append(res, key)
		}
	}
	return res
}

func redisConnect(host string, port int, db int) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: "",
		DB:       db,
	})
}

func main() {
	redisConnect("localhost", 6379, 0)
	sizes := map[string]int{}
	for _, list := range ListRedisLists() {
		size := rdb.LLen(ctx, list).Val()
		fmt.Printf("%s: %d\n", list, size)
		sizes[list] = int(size)
	}
}
