package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()

type Config struct {
	Host string
	Port int
	Db   int
}

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

func LoadOsDefaults(cfg *Config) {
	// Load environment variables if present, else set defaults.
	// Defaults are set to localhost:6379, db 0.
	// These can be overridden by passing in environment variables.
	// e.g. REDIS_HOST=localhost REDIS_PORT=6379 REDIS_DB=0 redis-llen-metrics
	// e.g. REDIS_HOST=localhost REDIS_PORT=6379 REDIS_DB=0 go run redis-llen-metrics.go -host=localhost -port=6379 -db=0
	host, present := os.LookupEnv("REDIS_HOST")
	if !present {
		cfg.Host = "localhost"
	} else {
		cfg.Host = host
	}
	port, present := os.LookupEnv("REDIS_PORT")
	if !present {
		cfg.Port = 6379
	} else {
		p, err := strconv.Atoi(port)
		if err != nil {
			log.Fatalf("Can't convert %v to int, review environment", port)
		}
		cfg.Port = p
	}
	db, present := os.LookupEnv("REDIS_PORT")
	if !present {
		cfg.Db = 0
	} else {
		d, err := strconv.Atoi(db)
		if err != nil {
			log.Fatalf("Can't convert %v to int, review environment", db)
		}
		cfg.Db = d
	}
}

func LoadFlagsDefaults(cfg *Config) {
	flag.StringVar(&cfg.Host, "host", cfg.Host, "Redis host")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Redis port")
	flag.IntVar(&cfg.Db, "db", cfg.Db, "Redis db")
	flag.Parse()
}

func ValidateConfig(cfg Config) error {
	if cfg.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if cfg.Port == 0 {
		return fmt.Errorf("redis port is required")
	}
	redisConnected := rdb.Ping(ctx).Val()
	if redisConnected != "PONG" {
		return fmt.Errorf("redis connection to %s:%d failed", cfg.Host, cfg.Port)
	}
	return nil
}

func main() {
	cfg := Config{}
	LoadOsDefaults(&cfg)
	LoadFlagsDefaults(&cfg)

	redisConnect(cfg.Host, cfg.Port, cfg.Db)

	err := ValidateConfig(cfg)
	if err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	sizes := map[string]int{}
	for _, list := range ListRedisLists() {
		size := rdb.LLen(ctx, list).Val()
		fmt.Printf("%s: %d\n", list, size)
		sizes[list] = int(size)
	}
}
