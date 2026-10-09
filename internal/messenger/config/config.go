package config

import "time"

type Config struct {
	PublicAddr  string        `env:"PUBLIC_HTTP_ADDR,notEmpty"`
	PostgresURL string        `env:"POSTGRES_URL,notEmpty"`
	JWTSecret   string        `env:"JWT_SECRET,notEmpty"`
	AccessTTL   time.Duration `env:"ACCESS_TTL,notEmpty"`
	RefreshTTL  time.Duration `env:"REFRESH_TTL,notEmpty"`
	NATSUrl     string        `env:"NATS_URL,notEmpty"`
	RedisAddr   string        `env:"REDIS_ADDR,notEmpty"`
	Argon2      Argon2        `envPrefix:"ARGON2_"`
}

type Argon2 struct {
	Memory      uint32 `env:"MEMORY,notEmpty"`
	Iterations  uint32 `env:"ITERATIONS,notEmpty"`
	Parallelism uint8  `env:"PARALLELISM,notEmpty"`
	SaltLength  uint32 `env:"SALT_LENGTH,notEmpty"`
	KeyLength   uint32 `env:"KEY_LENGTH,notEmpty"`
}
