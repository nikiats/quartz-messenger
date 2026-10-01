.PHONY: gen sqlc

gen:
	go generate ./...

sqlc:
	sqlc generate -f internal/messenger/repository/postgres/sqlc.yaml

build:
	go build ./...