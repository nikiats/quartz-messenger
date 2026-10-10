package config

import "time"

type Config struct {
	PublicAddr     string        `env:"PUBLIC_HTTP_ADDR,notEmpty"`
	PostgresURL    string        `env:"POSTGRES_URL,notEmpty"`
	Cassandra      Cassandra     `envPrefix:"CASSANDRA_"`
	JWTSecret      string        `env:"JWT_SECRET,notEmpty"`
	AccessTTL      time.Duration `env:"ACCESS_TTL,notEmpty"`
	RefreshTTL     time.Duration `env:"REFRESH_TTL,notEmpty"`
	NATSUrl        string        `env:"NATS_URL,notEmpty"`
	RedisAddr      string        `env:"REDIS_ADDR,notEmpty"`
	IdempotencyTTL time.Duration `env:"IDEMPOTENCY_TTL,notEmpty"`
	Argon2         Argon2        `envPrefix:"ARGON2_"`
}

type Cassandra struct {
	Hosts             []string `env:"HOSTS,notEmpty"`
	Keyspace          string   `env:"KEYSPACE,notEmpty"`
	DisableHostLookup bool     `env:"DISABLE_HOST_LOOKUP"`
}

type Argon2 struct {
	Memory      uint32 `env:"MEMORY,notEmpty"`
	Iterations  uint32 `env:"ITERATIONS,notEmpty"`
	Parallelism uint8  `env:"PARALLELISM,notEmpty"`
	SaltLength  uint32 `env:"SALT_LENGTH,notEmpty"`
	KeyLength   uint32 `env:"KEY_LENGTH,notEmpty"`
}
