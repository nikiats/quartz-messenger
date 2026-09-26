package main

import (
	"fmt"
	"os"

	"quaztz/internal/gateway/config"
	loader "quaztz/internal/platform/config"
)

func main() {
	var cfg config.Config
	if err := loader.Load(&cfg, "deploy/gateway.env"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
