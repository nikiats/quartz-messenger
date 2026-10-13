.PHONY: gen sqlc cqlgen build migrate-up migrate-down migrate-status migrate-create web-dev infra-up up down cqlsh

-include deploy/messenger.env

GOOSE = go run -tags=no_clickhouse,no_mssql,no_mysql,no_sqlite3,no_libsql,no_vertica,no_ydb \
	github.com/pressly/goose/v3/cmd/goose@v3.28.0 -dir migrations

gen:
	go generate ./...

sqlc:
	sqlc generate -f internal/messenger/repository/postgres/sqlc.yaml

cqlgen:
	docker run --rm --network quaztz_default -v "$(CURDIR)":/src -v quaztz-gomod:/go/pkg/mod -w /src golang:1.26-alpine \
		go tool schemagen -cluster=cassandra -keyspace=quaztz -output=internal/messenger/repository/cassandra/db -pkgname=db

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

web-dev:
	cd web && npm run dev

infra-up:
	docker compose up -d

up:
	docker compose --profile app up --build -d

down:
	docker compose --profile app --profile tools down

cqlsh:
	docker compose exec cassandra cqlsh
