package config

import "time"

type Config struct {
	WSAddr         string        `env:"WS_ADDR,notEmpty"`
	PrivateAddr    string        `env:"PRIVATE_HTTP_ADDR,notEmpty"`
	MessengerURL   string        `env:"MESSENGER_URL,notEmpty"`
	InternalToken  string        `env:"INTERNAL_TOKEN,notEmpty"`
	OriginPatterns []string      `env:"ORIGIN_PATTERNS,notEmpty" envSeparator:","`
	AccessTTL      time.Duration `env:"ACCESS_TTL,notEmpty"`
}
