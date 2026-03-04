.PHONY: migrate-up migrate-down migrate-down-all run-backend

# Migrations - run from project root; backend/migrations is used
migrate-up:
	cd backend && go run ./cmd/migrate

migrate-down:
	cd backend && go run ./cmd/migrate -down

migrate-down-all:
	cd backend && go run ./cmd/migrate -down-all

# Run backend (development)
run-backend:
	cd backend && go run ./cmd/api
