SHELL := /bin/sh
DB_NAME ?= racecontrol
DB_USER ?= racecontrol
DB_PASS ?= racecontrol
DB_PORT ?= 5432
GO ?= go

# Local configuration, excluded from version control. Every variable in it is
# exported so the Go binaries read it from the environment like in production.
-include .env
export

.PHONY: help db-up db-down db-logs db-ps migrate migrate-down db-reset seed sqlc backend-run backend-test backend-lint tidy app-install app-start app-web app-lint check

help: ## Show available targets
	@grep -hE '^[a-z-]+:.*?## ' $(MAKEFILE_LIST) | sed 's/:.*## /\t/' | expand -t16

# --- database ---------------------------------------------------------------

db-up: ## Start Postgres in a container (podman, also works with docker)
	-podman rm -f racecontrol-db >/dev/null 2>&1 || true
	podman run -d --name racecontrol-db \
		-e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASS) -e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 -v racecontrol-db-data:/var/lib/postgresql \
		--health-cmd "pg_isready -U $(DB_USER) -d $(DB_NAME)" \
		--health-interval 5s --health-retries 10 \
		docker.io/library/postgres:18-alpine

db-down: ## Stop and remove the Postgres container (keeps data)
	-podman rm -f racecontrol-db

db-logs: ## Tail Postgres logs
	podman logs -f racecontrol-db

db-ps: ## Show running containers
	podman ps

# --- backend ----------------------------------------------------------------

migrate: ## Apply database migrations
	cd backend && $(GO) run ./cmd/migrate

migrate-down: ## Roll back the last migration
	cd backend && $(GO) run ./cmd/migrate -down 1

db-reset: ## Drop the database, re-run migrations and re-seed from scratch
	-podman exec racecontrol-db psql -U $(DB_USER) -d postgres -q -c 'DROP DATABASE IF EXISTS $(DB_NAME) WITH (FORCE);'
	-podman exec racecontrol-db psql -U $(DB_USER) -d postgres -q -c 'CREATE DATABASE $(DB_NAME);'
	$(MAKE) migrate
	$(MAKE) seed

seed: ## Apply the idempotent reference data
	cd backend && $(GO) run ./cmd/seed

sqlc: ## Regenerate type-safe queries from SQL
	cd backend && sqlc generate

backend-run: ## Run the API locally
	cd backend && $(GO) run ./cmd/api

backend-test: ## Run backend tests
	cd backend && $(GO) test ./...

backend-lint: ## Vet and format-check the backend
	cd backend && $(GO) vet ./... && test -z "$$(gofmt -l .)"

tidy: ## Tidy Go module dependencies
	cd backend && $(GO) mod tidy

# --- app --------------------------------------------------------------------

app-install: ## Install app dependencies
	cd app && npm install

app-start: ## Start the Expo dev server
	cd app && npx expo start

app-web: ## Run the app in the browser (react-native-web)
	cd app && npx expo start --web

app-lint: ## Type-check the app
	@test -d app/node_modules || { echo "app dependencies not installed; run 'make app-install'"; exit 0; }; \
	cd app && npx tsc --noEmit

# --- everything -------------------------------------------------------------

check: backend-lint backend-test app-lint ## Lint and test everything