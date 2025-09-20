.PHONY: help build up down restart logs shell redis-cli psql clean dev tools

# Default target
help:
	@echo "Available commands:"
	@echo "  build     - Build Docker images"
	@echo "  up        - Start all services"
	@echo "  down      - Stop all services"
	@echo "  restart   - Restart all services"
	@echo "  logs      - View logs from all services"
	@echo "  logs-app  - View logs from the app service only"
	@echo "  shell     - Get a shell in the app container"
	@echo "  redis-cli - Connect to Redis CLI"
	@echo "  psql      - Connect to PostgreSQL"
	@echo "  clean     - Remove all containers, volumes, and images"
	@echo "  dev       - Start services in development mode"
	@echo "  tools     - Start with admin tools (pgAdmin, Redis Commander)"
	@echo "  test      - Run tests"

# Build Docker images
build:
	docker compose build

# Start all services
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# Restart all services
restart: down up

# View logs
logs:
	docker compose logs -f

# View app logs only
logs-app:
	docker compose logs -f app

# Get a shell in the app container
shell:
	docker compose exec app sh

# Connect to Redis CLI
redis-cli:
	docker compose exec redis redis-cli

# Connect to PostgreSQL
psql:
	docker compose exec postgres psql -U ${DB_USER:-postgres} -d ${DB_NAME:-matching_db}

# Clean up everything
clean:
	docker compose down -v --rmi all --remove-orphans

# Development mode (build and run)
dev: build up logs

# Start with admin tools
tools:
	docker compose --profile tools up -d

# Run tests
test:
	go test -v ./...

# Generate Swagger docs
swagger:
	swag init -g cmd/server/main.go

# Run locally (without Docker)
run:
	go run cmd/server/main.go

# Hot reload for development
watch:
	air