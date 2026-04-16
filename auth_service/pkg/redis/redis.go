package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	URL          string
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PingTimeout  time.Duration
}

func DefaultConfig() Config {
	return Config{
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PingTimeout:  3 * time.Second,
	}
}

func New(ctx context.Context, cfg Config) (*redis.Client, error) {
	defaults := DefaultConfig()
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = defaults.DialTimeout
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaults.ReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = defaults.WriteTimeout
	}
	if cfg.PingTimeout == 0 {
		cfg.PingTimeout = defaults.PingTimeout
	}

	var opts *redis.Options
	if cfg.URL != "" {
		var err error
		opts, err = redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("redis: parse url: %w", err)
		}
	} else {
		if cfg.Addr == "" {
			return nil, fmt.Errorf("redis: URL or Addr is required")
		}
		opts = &redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}
	}
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout

	client := redis.NewClient(opts)
	pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}
	return client, nil
}

func NewFromURL(ctx context.Context, url string) (*redis.Client, error) {
	cfg := DefaultConfig()
	cfg.URL = url
	return New(ctx, cfg)
}
