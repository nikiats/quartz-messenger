.PHONY: gen sqlc migrate-up migrate-down migrate-status migrate-create

-include deploy/messenger.env

GOOSE = go run -tags=no_clickhouse,no_mssql,no_mysql,no_sqlite3,no_libsql,no_vertica,no_ydb \
	github.com/pressly/goose/v3/cmd/goose@v3.28.0 -dir migrations

gen:
	go generate ./...

sqlc:
	sqlc generate -f internal/messenger/repository/postgres/sqlc.yaml

build:
	go build ./...

migrate-up:
	$(GOOSE) postgres "$(POSTGRES_URL)" up

migrate-down:
	$(GOOSE) postgres "$(POSTGRES_URL)" down

migrate-status:
	$(GOOSE) postgres "$(POSTGRES_URL)" status

migrate-create:
	$(GOOSE) -s create $(name) sql
