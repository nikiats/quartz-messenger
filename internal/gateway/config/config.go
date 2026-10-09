package config

type Config struct {
	WSAddress      string   `env:"WS_ADDR,notEmpty"`
	NATSUrl        string   `env:"NATS_URL,notEmpty"`
	RedisAddress   string   `env:"REDIS_ADDR,notEmpty"`
	JWTSecret      string   `env:"JWT_SECRET,notEmpty"`
	OriginPatterns []string `env:"ORIGIN_PATTERNS,notEmpty" envSeparator:","`
}
