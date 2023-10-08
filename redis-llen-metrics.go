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
	Host       string
	Port       int
	Db         int
	LoopMode   bool
	LoopTime   int
	DumpConfig bool
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

func OsEnvWithDefault(env, defaultValue string) string {
	value, present := os.LookupEnv(env)
	if present {
		return value
	}
	return defaultValue
}

func LoadOsDefaults(cfg *Config) {
	// Load environment variables if present, else set defaults.
	// Defaults are set to localhost:6379, db 0.
	// These can be overridden by passing in environment variables.
	// e.g. REDIS_HOST=localhost REDIS_PORT=6379 REDIS_DB=0 redis-llen-metrics
	// e.g. REDIS_HOST=localhost REDIS_PORT=6379 REDIS_DB=0 go run redis-llen-metrics.go -host=localhost -port=6379 -db=0

	cfg.Host = OsEnvWithDefault("REDIS_HOST", "localhost")
	port := OsEnvWithDefault("REDIS_PORT", "6379")
	db := OsEnvWithDefault("REDIS_DB", "0")

	p, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Can't convert %v to int, review environment", port)
	}
	cfg.Port = p
	d, err := strconv.Atoi(db)
	if err != nil {
		log.Fatalf("Can't convert %v to int, review environment", db)
	}
	cfg.Db = d

	cfg.LoopMode = false
	cfg.LoopTime = 0
}

func LoadFlagsDefaults(cfg *Config) {
	flag.StringVar(&cfg.Host, "host", cfg.Host, "Redis host")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Redis port")
	flag.IntVar(&cfg.Db, "db", cfg.Db, "Redis db")
	flag.IntVar(&cfg.LoopTime, "loop-time", cfg.LoopTime, "Loop time in seconds")
	flag.BoolVar(&cfg.DumpConfig, "config", false, "Dump config")
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
	backends := []Backend{}
	backends = append(backends, new(TextBackend))
	for _, backend := range backends {
		backend.AddFlags()
	}
	LoadOsDefaults(&cfg)
	LoadFlagsDefaults(&cfg)
	if cfg.DumpConfig {
		fmt.Printf("Config: %+v\n", cfg)
		for _, backend := range backends {
			fmt.Printf("Backend: %+v\n", backend)
		}
		os.Exit(0)
	}

	for _, backend := range backends {
		backend.Activate()
	}

	for _, backend := range backends {
		defer backend.Close()
	}

	redisConnect(cfg.Host, cfg.Port, cfg.Db)

	err := ValidateConfig(cfg)
	if err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	for cfg.LoopMode {
		metricMemory.ResetMetrics()
		for _, list := range ListRedisLists() {
			size := rdb.LLen(ctx, list).Val()
			metricMemory.AddMetric(Metric{list, int(size)})
		}

		for _, backend := range backends {
			backend.PublishStats()
		}

		time.Sleep(time.Duration(cfg.LoopTime) * time.Second)
	}
}
