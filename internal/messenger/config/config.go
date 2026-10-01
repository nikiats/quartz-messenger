package config

import "time"

type Config struct {
	PublicAddr          string        `env:"PUBLIC_HTTP_ADDR,notEmpty"`
	PrivateAddr         string        `env:"PRIVATE_HTTP_ADDR,notEmpty"`
	PostgresURL         string        `env:"POSTGRES_URL,notEmpty"`
	JWTSecret           string        `env:"JWT_SECRET,notEmpty"`
	AccessTTL           time.Duration `env:"ACCESS_TTL,notEmpty"`
	RefreshTTL          time.Duration `env:"REFRESH_TTL,notEmpty"`
	InternalToken       string        `env:"INTERNAL_TOKEN,notEmpty"`
	GatewayURL          string        `env:"GATEWAY_URL,notEmpty"`
	MaxGatewayIdleConns int           `env:"GATEWAY_MAX_IDLE_CONNS,notEmpty"`
}
