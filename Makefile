.PHONY: help up down restart logs build clean test generate-keys

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## View logs from all services
	docker-compose logs -f

logs-go: ## View logs from Go service
	docker-compose logs -f go-app

logs-xray: ## View logs from Xray service
	docker-compose logs -f xray

build: ## Build all services
	docker-compose build

rebuild: ## Rebuild and restart all services
	docker-compose build --no-cache
	docker-compose up -d

clean: ## Stop services and remove volumes
	docker-compose down -v

status: ## Show status of all services
	docker-compose ps

generate-keys: ## Generate Xray Reality keys
	@echo "Generating Xray Reality keys..."
	@docker run --rm teddysun/xray:latest xray x25519

test-api: ## Test API endpoints
	@echo "Testing health endpoint..."
	@curl -s http://localhost:8080/health | jq .
	@echo "\nTesting metrics endpoint..."
	@curl -s http://localhost:8080/metrics | head -n 20

create-user: ## Create a test user (requires email and uuid params)
	@curl -X POST http://localhost:8080/api/users/create \
		-H "Content-Type: application/json" \
		-d '{"email":"test@example.com","uuid":"b831381d-6324-4d53-ad4f-8cda48b30811"}' | jq .

list-users: ## List all users
	@curl -s http://localhost:8080/api/users/list | jq .

open-grafana: ## Open Grafana in browser
	@open http://localhost:3000 || xdg-open http://localhost:3000

open-prometheus: ## Open Prometheus in browser
	@open http://localhost:9090 || xdg-open http://localhost:9090

install-dev: ## Install development tools
	@echo "Installing development tools..."
	@command -v jq >/dev/null 2>&1 || echo "Please install jq: brew install jq"
	@command -v docker >/dev/null 2>&1 || echo "Please install Docker"
	@command -v docker-compose >/dev/null 2>&1 || echo "Please install Docker Compose"

