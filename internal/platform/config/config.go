package config

import (
	"errors"
	"io/fs"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func Load[T any](cfg *T, envFile string) error {
	if err := godotenv.Load(envFile); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return env.Parse(cfg)
}
