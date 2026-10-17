.PHONY: gen cqlgen build migrate-create web-dev infra-up up down cqlsh

gen:
	go generate ./...

cqlgen:
	docker run --rm --network quartz_default -v "$(CURDIR)":/src -w /src golang:1.26-alpine \
		go tool schemagen -cluster=cassandra -keyspace=quartz_messenger -output=internal/messenger/repository/cassandra/db -pkgname=db

build:
	go build ./...

migrate-create:
	go run github.com/pressly/goose/v3/cmd/goose@v3.28.0 -dir migrations -s create $(name) sql

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
