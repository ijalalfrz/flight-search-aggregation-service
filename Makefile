#===================#
#== Env Variables ==#
#===================#
DOCKER_COMPOSE_FILE ?= docker-compose.dev.yml
WITH_TTY ?= -t

SWAG := $(shell go env GOPATH)/bin/swag

api-docs: ## Generate API docs with swaggo
	@echo "========================="
	@echo "Generate Swagger API Docs"
	@echo "========================="
	go install github.com/swaggo/swag/cmd/swag@latest
	$(SWAG) init --parseDependency --parseInternal -g ./cmd/main.go -ot "json" -o ./docs --instanceName flight-search-aggregation


start:
	@echo "========================="
	@echo "Restarting services..."
	@echo "========================="
	docker compose -f ${DOCKER_COMPOSE_FILE} up -d --build
	docker compose -f ${DOCKER_COMPOSE_FILE} ps


stop:
	@echo "========================="
	@echo "Stopping services..."
	@echo "========================="
	docker compose -f ${DOCKER_COMPOSE_FILE} down --remove-orphans
	docker compose -f ${DOCKER_COMPOSE_FILE} ps


restart: stop start

setup-env:
	@echo "========================="
	@echo "Setting up environment..."
	@echo "========================="
	cp ./.env.example ./.env

setup: setup-env start

tests-unit:
	go test -tags=unit -v -timeout 10s -count=1 ./... -coverprofile=coverage.out

tests-load:
	APP_HOST=http://localhost:8080 REDIS_ADDR=localhost:6379 REDIS_PASSWORD=redis123 go test -tags=load -v -count=2 ./tests/load/...