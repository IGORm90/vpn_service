.PHONY: help up down restart logs build clean generate-keys setup env-check

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: generate-keys ## Setup project (generate keys and create .env)
	@echo "Setting up VPN service..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file. Please edit it with your settings."; \
	else \
		echo ".env file already exists"; \
	fi

env-check: ## Check if .env is configured
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Run 'make setup' first."; \
		exit 1; \
	fi
	@if grep -q "YOUR_PRIVATE_KEY_HERE" .env; then \
		echo "Error: Please set XRAY_PRIVATE_KEY in .env file"; \
		exit 1; \
	fi
	@if grep -q "YOUR_SERVER_IP_HERE" .env; then \
		echo "Warning: Please set SERVER_IP in .env file"; \
	fi

up: env-check ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## View logs from all services
	docker-compose logs -f

logs-go: ## View logs from Go service
	docker-compose logs -f app

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
	@echo "Copy the 'Private key' to XRAY_PRIVATE_KEY in .env file"
	@echo ""
	@docker run --rm teddysun/xray:latest xray x25519

# API Testing
test-health: ## Test health endpoint
	@echo "Testing health endpoint..."
	@curl -s http://localhost:8080/health | jq .

test-stats: ## Test stats endpoint
	@echo "Testing stats endpoint..."
	@curl -s http://localhost:8080/stats | jq .

# User Management
create-user: ## Create a test user
	@echo "Creating test user..."
	@curl -X POST http://localhost:8080/api/users \
		-H "Content-Type: application/json" \
		-d '{"username":"testuser","password":"testpass123","traffic_limit":10737418240}' | jq .

list-users: ## List all users
	@echo "Listing all users..."
	@curl -s http://localhost:8080/api/users | jq .

get-user-config: ## Get user config (requires USER_ID env var)
	@if [ -z "$(USER_ID)" ]; then \
		echo "Error: Please provide USER_ID, e.g., make get-user-config USER_ID=1"; \
		exit 1; \
	fi
	@echo "Getting config for user $(USER_ID)..."
	@curl -s http://localhost:8080/api/users/$(USER_ID)/config | jq .

delete-user: ## Delete user (requires USER_ID env var)
	@if [ -z "$(USER_ID)" ]; then \
		echo "Error: Please provide USER_ID, e.g., make delete-user USER_ID=1"; \
		exit 1; \
	fi
	@echo "Deleting user $(USER_ID)..."
	@curl -X DELETE http://localhost:8080/api/users/$(USER_ID)

# Monitoring
open-grafana: ## Open Grafana in browser
	@open http://localhost:3000 || xdg-open http://localhost:3000

open-prometheus: ## Open Prometheus in browser
	@open http://localhost:9090 || xdg-open http://localhost:9090

metrics: ## Show Prometheus metrics
	@curl -s http://localhost:8080/metrics | grep ^vpn

# Development
install-dev: ## Install development tools
	@echo "Installing development tools..."
	@command -v jq >/dev/null 2>&1 || echo "Please install jq: brew install jq"
	@command -v docker >/dev/null 2>&1 || echo "Please install Docker"
	@command -v docker-compose >/dev/null 2>&1 || echo "Please install Docker Compose"

db-backup: ## Backup SQLite database
	@echo "Backing up database..."
	@docker cp vpn-app:/app/data/vpn.db ./vpn.db.backup
	@echo "Database backed up to vpn.db.backup"

db-restore: ## Restore SQLite database from backup
	@if [ ! -f vpn.db.backup ]; then \
		echo "Error: vpn.db.backup not found"; \
		exit 1; \
	fi
	@echo "Restoring database..."
	@docker cp ./vpn.db.backup vpn-app:/app/data/vpn.db
	@docker-compose restart app
	@echo "Database restored"

