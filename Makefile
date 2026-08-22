APP_NAME := xlink
BIN_DIR := bin
MIGRATIONS_DIR := migrations

ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: run build test migrate-up migrate-down migrate-create

run:
	go run ./cmd/api

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

test:
	go test -v -race ./...

migrate-up:
	@test -n "$(DB_URL)" || (echo "DB_URL is required (set it in .env or pass it to make)"; exit 1)
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	@test -n "$(DB_URL)" || (echo "DB_URL is required (set it in .env or pass it to make)"; exit 1)
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-create:
	@test -n "$(name)" || (echo "Migration name is required. Usage: make migrate-create name=create_example_table"; exit 1)
	go run github.com/golang-migrate/migrate/v4/cmd/migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)
