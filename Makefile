COMPOSE = docker compose --env-file backend/.env
.PHONY: run docker-up docker-down fmt test vet check

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
