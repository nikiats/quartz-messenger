package config

import "time"

type Config struct {
	WSAddress             string        `env:"WS_ADDR,notEmpty"`
	NATSUrl               string        `env:"NATS_URL,notEmpty"`
	RedisAddress          string        `env:"REDIS_ADDR,notEmpty"`
	OriginPatterns        []string      `env:"ORIGIN_PATTERNS,notEmpty" envSeparator:","`
	AccessTTL             time.Duration `env:"ACCESS_TTL,notEmpty"`
	MaxMessengerIdleConns int           `env:"MESSENGER_MAX_IDLE_CONNS,notEmpty"`
}
