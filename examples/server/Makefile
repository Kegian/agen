PG_USER ?= postgres
PG_PASS ?= postgres
PG_HOST ?= localhost
PG_PORT ?= 5432
PG_NAME ?= postgres

CH_USER ?= default
CH_PASS ?= default
CH_HOST ?= localhost
CH_PORT ?= 9000
CH_NAME ?= default

build-docker:
	docker-compose build

build:
	go build -o out/server main.go

run-docker:
	docker-compose up

run:
	go run main.go

install:
	go install -v github.com/Kegian/agen/cmd/agen@latest
	go install -v github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -v github.com/ogen-go/ogen/cmd/ogen@latest
	go install -v -tags 'postgres clickhouse' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

gen-api:
	agen gen -i api.yml -o internal/generated

gen-sql:
	sqlc generate

gen: gen-api gen-sql

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

migrate-postgres:
	migrate -source file://migrations/postgres -database "postgres://$(PG_USER):$(PG_PASS)@$(PG_HOST):$(PG_PORT)/$(PG_NAME)?sslmode=disable&x-multi-statement=true" up

migrate-clickhouse:
	migrate -source file://migrations/clickhouse -database "clickhouse://$(CH_HOST):$(CH_PORT)?username=$(CH_USER)&password=$(CH_PASS)&database=$(CH_NAME)&x-multi-statement=true" up

migrate: migrate-postgres migrate-clickhouse

lint:
	docker run -t --rm -v ${PWD}:/app -v ~/.cache/golangci-lint/v1.55.2:/root/.cache -w /app golangci/golangci-lint:v1.55.2 golangci-lint run -v

lint-clear-cache:
	rm -rf ~/.cache/golangci-lint/v1.55.2:/root/.cache
