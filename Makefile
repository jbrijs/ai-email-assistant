.PHONY: help up down build logs clean migrate migrate-down migrate-status create-migration

help: ## Show this help message
	@echo "InboxAI - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Start all services
	docker compose up -d

down: ## Stop all services
	docker compose down

build: ## Build and start services
	docker compose up -d --build

logs: ## Show logs for all services
	docker compose logs -f

clean: ## Stop and remove all containers, networks, and volumes
	docker compose down -v --remove-orphans

migrate: ## Run all pending migrations
	docker compose run --rm migrations -path /migrations -database postgres://postgres:postgres@db:5432/inboxai?sslmode=disable up

migrate-down: ## Rollback last migration
	docker compose run --rm migrations -path /migrations -database postgres://postgres:postgres@db:5432/inboxai?sslmode=disable down 1

migrate-status: ## Check migration status
	docker compose run --rm migrations -path /migrations -database postgres://postgres:postgres@db:5432/inboxai?sslmode=disable version

create-migration: ## Create a new migration file (usage: make create-migration NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make create-migration NAME=migration_name"; \
		echo "Example: make create-migration NAME=add_user_preferences"; \
		exit 1; \
	fi
	cd db && ./create-migration.sh $(NAME)

db-shell: ## Connect to database shell
	docker compose exec db psql -U postgres -d inboxai

api-shell: ## Connect to API container shell
	docker compose exec api sh
