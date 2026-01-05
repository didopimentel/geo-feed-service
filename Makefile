# ============================
# Configuration
# ============================

COMPOSE := docker compose
PROJECT := geo-feed-service

POSTGRES_URL := postgres://feed_user:feed_password@localhost:5432/feed?sslmode=disable
REDIS_URL := redis://localhost:6379

GO := go
GOFLAGS := -v

# ============================
# Docker
# ============================

.PHONY: up
up:
	$(COMPOSE) up -d

.PHONY: down
down:
	$(COMPOSE) down

.PHONY: restart
restart: down up

.PHONY: logs
logs:
	$(COMPOSE) logs -f

.PHONY: ps
ps:
	$(COMPOSE) ps

# ============================
# Database
# ============================

.PHONY: db-shell
db-shell:
	$(COMPOSE) exec postgres psql -U feed_user -d feed

.PHONY: redis-shell
redis-shell:
	$(COMPOSE) exec redis redis-cli

.PHONY: db-reset
db-reset:
	$(COMPOSE) down -v
	$(COMPOSE) up -d

# ============================
# Go App
# ============================

.PHONY: run
run:
	POSTGRES_URL=$(POSTGRES_URL) \
	REDIS_URL=$(REDIS_URL) \
	$(GO) run ./cmd/api

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o bin/api ./cmd/api

.PHONY: test
test:
	$(GO) test ./... -count=1

.PHONY: test-integration
test-integration:
	POSTGRES_URL=$(POSTGRES_URL) \
	REDIS_URL=$(REDIS_URL) \
	$(GO) test ./... -tags=integration -count=1

# ============================
# Quality
# ============================

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: lint
lint:
	golangci-lint run

# ============================
# Migrations (optional)
# ============================

.PHONY: migrate
migrate:
	migrate -database "$(POSTGRES_URL)" -path migrations up

.PHONY: migrate-down
migrate-down:
	migrate -database "$(POSTGRES_URL)" -path migrations down

# ============================
# Help
# ============================

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  up               Start Postgres + Redis"
	@echo "  down             Stop containers"
	@echo "  restart          Restart containers"
	@echo "  logs             Tail container logs"
	@echo "  ps               List containers"
	@echo ""
	@echo "  run              Run API locally"
	@echo "  build            Build API binary"
	@echo "  test             Run unit tests"
	@echo "  test-integration Run integration tests"
	@echo ""
	@echo "  db-shell         Open psql shell"
	@echo "  redis-shell      Open redis-cli"
	@echo "  db-reset         Reset DB volumes"
	@echo ""
	@echo "  fmt              Go fmt"
	@echo "  vet              Go vet"
	@echo "  lint             GolangCI lint"
	@echo ""
	@echo "  migrate          Run DB migrations"
	@echo "  migrate-down     Rollback migrations"
