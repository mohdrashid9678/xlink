APP_NAME := xlink
BIN_DIR := bin
SCHEMA_FILE := migrations/init_schema.sql

# Let `make db-init` use DB_URL from a local .env file when it exists.
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: run build test db-init

run:
	go run ./cmd/api

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/api

test:
	go test ./...

# Applies the fixed initial schema. It intentionally fails if the table exists.
db-init:
	@test -n "$(DB_URL)" || (echo "DB_URL is required (set it in .env or pass it to make)"; exit 1)
	psql "$(DB_URL)" -v ON_ERROR_STOP=1 -f $(SCHEMA_FILE)
