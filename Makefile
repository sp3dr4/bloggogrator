.DEFAULT_GOAL := help
SHELL = bash

## help: print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test-cover: run all tests and display coverage
.PHONY: test-cover
test-cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## psql: opens the psql cli
.PHONY: psql
psql:
	docker compose exec pg psql -U admin -d test_db

.PHONY: _migrate
_migrate:
	goose -v -dir ./sql/schema postgres "postgres://admin:admin@localhost:5432/test_db" ${COMMAND}

## migrate-status: shows the migrations status
.PHONY: migrate-status
migrate-status:
	@$(MAKE) _migrate COMMAND="status"

## migrate-up: applies all the migrations
.PHONY: migrate-up
migrate-up:
	@$(MAKE) _migrate COMMAND="up"

## migrate-reset: undoes all the migrations
.PHONY: migrate-reset
migrate-reset:
	@$(MAKE) _migrate COMMAND="reset"

## sqlc-gen: generates code from sql files
.PHONY: sqlc-gen
sqlc-gen:
	@sqlc generate
	@echo "done"
