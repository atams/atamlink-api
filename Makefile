# Variables
APP_NAME=catalogd
MAIN_PATH=cmd/catalogd/main.go
DOCKER_IMAGE=atamlink-catalog
DB_CONTAINER=atamlink-postgres
REDIS_CONTAINER=atamlink-redis

# Go commands
.PHONY: run
run: ## Run aplikasi
	go run $(MAIN_PATH)

.PHONY: build
build: ## Build aplikasi
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests dengan coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: tidy
tidy: ## Tidy dan download dependencies
	go mod tidy
	go mod download

# Database commands
.PHONY: db-up
db-up: ## Start database container
	docker run --name $(DB_CONTAINER) \
		-e POSTGRES_USER=atamlink_user \
		-e POSTGRES_PASSWORD=atamlink_password \
		-e POSTGRES_DB=atamlink_db \
		-p 5432:5432 \
		-d postgres:15-alpine

.PHONY: db-down
db-down: ## Stop database container
	docker stop $(DB_CONTAINER)
	docker rm $(DB_CONTAINER)

.PHONY: db-logs
db-logs: ## Show database logs
	docker logs -f $(DB_CONTAINER)

.PHONY: db-shell
db-shell: ## Access database shell
	docker exec -it $(DB_CONTAINER) psql -U atamlink_user -d atamlink_db

# Migration commands
.PHONY: migrate-up
migrate-up: ## Run database migrations up
	migrate -path internal/database/migrations \
		-database "postgresql://atamlink_user:atamlink_password@localhost:5432/atamlink_db?sslmode=disable" \
		up

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	migrate -path internal/database/migrations \
		-database "postgresql://atamlink_user:atamlink_password@localhost:5432/atamlink_db?sslmode=disable" \
		down

.PHONY: migrate-create
migrate-create: ## Create migration file (usage: make migrate-create name=create_users_table)
	migrate create -ext sql -dir internal/database/migrations -seq $(name)

# Docker commands
.PHONY: docker-build
docker-build: ## Build docker image
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run: ## Run docker container
	docker run --name $(APP_NAME) \
		--network host \
		-e DB_HOST=localhost \
		-p 8080:8080 \
		-d $(DOCKER_IMAGE)

.PHONY: docker-stop
docker-stop: ## Stop docker container
	docker stop $(APP_NAME)
	docker rm $(APP_NAME)

# Development commands
.PHONY: dev
dev: ## Run dengan air untuk hot reload
	air

.PHONY: dev-setup
dev-setup: ## Setup development environment
	go install github.com/cosmtrek/air@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	cp .env.example .env
	@echo "Development setup complete!"

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

# Utility commands
.PHONY: check-env
check-env: ## Check environment variables
	@echo "Checking environment variables..."
	@test -f .env || (echo "Error: .env file not found" && exit 1)
	@echo "Environment file found!"

.PHONY: gen-docs
gen-docs: ## Generate API documentation
	swag init -g $(MAIN_PATH) -o ./docs

.PHONY: help
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help