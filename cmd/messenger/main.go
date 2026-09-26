package main

import (
	"fmt"
	"os"

	"quaztz/internal/messenger/config"
	loader "quaztz/internal/platform/config"
)

func main() {
	var cfg config.Config
	if err := loader.Load(&cfg, "deploy/messenger.env"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
