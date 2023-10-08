package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()
var metricMemory = MetricsStorage{}

type Config struct {
	Host     string
	Port     int
	Db       int
	LoopMode bool
	LoopTime int
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
	cfg.LoopMode = false
	cfg.LoopTime = 0
}

func LoadFlagsDefaults(cfg *Config) {
	flag.StringVar(&cfg.Host, "host", cfg.Host, "Redis host")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Redis port")
	flag.IntVar(&cfg.Db, "db", cfg.Db, "Redis db")
	flag.IntVar(&cfg.LoopTime, "loop-time", cfg.LoopTime, "Loop time in seconds")
	flag.Parse()

	if cfg.LoopTime > 0 {
		cfg.LoopMode = true
	}
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
	if cfg.LoopMode && cfg.LoopTime < 10 {
		return fmt.Errorf("loop time must be greater than 10 seconds")
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
	fmt.Printf("Debug config %v\n", cfg)

	for cfg.LoopMode {
		metricMemory.ResetMetrics()
		for _, list := range ListRedisLists() {
			size := rdb.LLen(ctx, list).Val()
			metricMemory.AddMetric(Metric{list, int(size)})
		}

		for _, metric := range metricMemory.Metrics {
			fmt.Printf("%s %d\n", metric.Name, metric.Value)
		}

		time.Sleep(time.Duration(cfg.LoopTime) * time.Second)
	}
}
