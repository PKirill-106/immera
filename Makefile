.PHONY: run docker-up docker-down fmt test vet check

run:
	cd backend && go run ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down

fmt:
	cd backend && gofmt -w $$(find . -name '*.go' -type f)

test:
	cd backend && go test ./...

vet:
	cd backend && go vet ./...

check: fmt test vet
