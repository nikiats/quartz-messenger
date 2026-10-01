// Package migrations содержит миграции схемы базы данных мессенджера в формате goose.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
