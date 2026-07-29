COMPOSE = docker compose --env-file backend/.env
ENV_FILE ?= backend/.env
TOOLS_DIR := $(CURDIR)/bin
GOOSE_VERSION := v3.27.1
GOOSE := $(TOOLS_DIR)/goose
MIGRATIONS_DIR := backend/migrations
GOOSE_ENV = GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DATABASE_URL)" GOOSE_MIGRATION_DIR="$(MIGRATIONS_DIR)"

ifeq ($(origin DATABASE_URL), undefined)
-include $(ENV_FILE)
endif

.PHONY: go-run run docker-up docker-down docker-down-v fmt test vet check \
	migrate-install migrate-create migrate-validate migrate-status \
	migrate-version migrate-up migrate-down check-database-url

go-run:
	cd backend && go run ./cmd/api

run:
	$(COMPOSE) up -d postgres 
	make go-run

docker-up:
	$(COMPOSE) up --build

docker-down:
	$(COMPOSE) down

docker-down-v:
	$(COMPOSE) down -v

fmt:
	cd backend && gofmt -w $$(find . -name '*.go' -type f)

test:
	cd backend && go test ./...

vet:
	cd backend && go vet ./...

check: fmt test vet

$(GOOSE):
	mkdir -p $(TOOLS_DIR)
	GOBIN=$(TOOLS_DIR) go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

migrate-install: $(GOOSE)

migrate-create: $(GOOSE)
	@test -n "$(name)" || (echo "usage: make migrate-create name=create_documents"; exit 1)
	@$(GOOSE) -dir "$(MIGRATIONS_DIR)" create "$(name)" sql

migrate-validate: $(GOOSE)
	@$(GOOSE) -dir "$(MIGRATIONS_DIR)" validate

check-database-url:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required; export it or set it in $(ENV_FILE)"; exit 1)

migrate-status: check-database-url $(GOOSE)
	@$(GOOSE_ENV) $(GOOSE) status

migrate-version: check-database-url $(GOOSE)
	@$(GOOSE_ENV) $(GOOSE) version

migrate-up: check-database-url $(GOOSE)
	@$(GOOSE_ENV) $(GOOSE) up

migrate-down: check-database-url $(GOOSE)
	@$(GOOSE_ENV) $(GOOSE) down
