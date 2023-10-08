package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cactus/go-statsd-client/v5/statsd"
)

type StatsDBackend struct {
	Host     string
	Port     int
	IsActive bool
	Prefix   string
	Suffix   string
	Client   statsd.Statter
}

func (s *StatsDBackend) AddFlags() {
	s.Host = OsEnvWithDefault("RLM_STATSD_HOST", "127.0.0.1")
	sport := OsEnvWithDefault("RLM_STATSD_PORT", "8125")
	OsEnvWithDefault("RLM_TEXT_SEPARATOR", "\t")
	OsEnvWithDefault("RLM_TEXT_PREFIX", "")
	s.Prefix = OsEnvWithDefault("RLM_STATSD_PREFIX", "redis.llen")
	s.Suffix = OsEnvWithDefault("RLM_STATSD_SUFFIX", "")

	port, err := strconv.Atoi(sport)
	if err != nil {
		log.Println("Error specified port is not a number, check environment")
		port = 0
	}

	flag.BoolVar(&s.IsActive, "statsd", false, "Activate StatsD backend")
	flag.StringVar(&s.Host, "statsd-host", s.Host, "StatsD host")
	flag.StringVar(&s.Prefix, "statsd-prefix", s.Prefix, "StatsD prefix")
	flag.StringVar(&s.Suffix, "statsd-suffix", s.Suffix, "StatsD suffix")
	flag.IntVar(&s.Port, "statsd-port", port, "StatsD port")
}

func (s *StatsDBackend) Activate() {
	if !s.IsActive {
		return
	}
	if s.Port == 0 {
		log.Fatalf("Error specified statsd port %v is not supported, check config", s.Port)
		os.Exit(1)
	}

	config := &statsd.ClientConfig{
		Address:     fmt.Sprintf("%v:%v", s.Host, s.Port),
		Prefix:      s.Prefix,
		UseBuffered: false,
	}

	client, err := statsd.NewClientWithConfig(config)
	if err != nil {
		log.Fatalf("Error creating statsd client: %v", err)
		os.Exit(1)
	}
	s.Client = client
}

func (s *StatsDBackend) Close() {
	if !s.IsActive {
		return
	}
	s.Client.Close()
}

func (s *StatsDBackend) PublishStats() error {
	if !s.IsActive {
		return nil
	}
	log.Printf("calling Statsd")
	for _, stat := range metricMemory.Metrics {
		s.Client.Gauge(stat.Name+s.Suffix, int64(stat.Value), 1.0)
	}
	return nil
}
