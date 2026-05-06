# Debtease – migration and integration tests
# Prereqs: Docker (for Postgres), or a running Postgres. For full migrations use goose.

DB_URL ?= postgres://postgres:postgres@localhost:5432/debtease?sslmode=disable

# Start Postgres (docker-compose)
up:
	docker compose up -d postgres

# Run only the finance_steps migration (Go; no goose needed). Requires calculation_sessions to exist.
migrate-finance-steps:
	@if [ -z "$$DB_URL" ]; then export DB_URL=$(DB_URL); fi; \
	go run scripts/migrate_finance_steps.go

# Run all migrations with goose (install: go install github.com/pressly/goose/v3/cmd/goose@latest)
migrate:
	goose -dir database/schema postgres "$(DB_URL)" up

# Run integration tests (require DB_URL and migrated DB)
test-integration:
	DB_URL="$(DB_URL)" go test ./cmd -tags=integration -v -count=1

# One-shot: start Postgres, run finance_steps migration, run integration tests.
# Use this if you already have calculation_sessions (e.g. from a previous full migrate).
integration: up
	@echo "Waiting for Postgres..."
	@sleep 5
	@$(MAKE) migrate-finance-steps
	@$(MAKE) test-integration

.PHONY: up migrate migrate-finance-steps test-integration integration
