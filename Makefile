.PHONY: dev install lint test build migrate pr docs stop

# ───────────────────────────────────────────
# Setup
# ───────────────────────────────────────────
install:
	@echo "Installing Go dev tools..."
	cd api && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	cd api && go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	cd api && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Installing docs dependencies..."
	cd docs && npm install
	@echo "Done. Run 'make dev' to start."

# ───────────────────────────────────────────
# Development
# ───────────────────────────────────────────
dev:
	docker-compose up -d postgres redis
	@echo ""
	@echo "Infrastructure running:"
	@echo "  Postgres: localhost:5432"
	@echo "  Redis:    localhost:6379"
	@echo ""
	@echo "Run these in separate terminals:"
	@echo "  API:      cd api && go run cmd/api/main.go"
	@echo "  Worker:   cd api && go run cmd/worker/main.go"
	@echo "  MCP:      cd api && go run cmd/mcp-server/main.go"
	@echo "  Web:      cd frontend && npm run dev"
	@echo "  Docs:     cd docs && npx mintlify dev"

stop:
	docker-compose down

# ───────────────────────────────────────────
# Database
# ───────────────────────────────────────────
migrate-up:
	cd api && migrate -path db/migrations -database "postgres://signal:signal@localhost:5432/signal?sslmode=disable" up

migrate-down:
	cd api && migrate -path db/migrations -database "postgres://signal:signal@localhost:5432/signal?sslmode=disable" down 1

migrate-create:
	@read -p "Migration name: " name; \
	cd api && migrate create -ext sql -dir db/migrations -seq $$name

sqlc:
	cd api && sqlc generate

# ───────────────────────────────────────────
# Lint & Test
# ───────────────────────────────────────────
lint:
	cd api && golangci-lint run ./...
	cd frontend && npm run lint

test:
	cd api && go test -v -race ./...
	cd frontend && npm run test

# ───────────────────────────────────────────
# Build
# ───────────────────────────────────────────
build:
	cd api && docker build -t signal-api .
	cd api && docker build -f Dockerfile.worker -t signal-worker .
	cd api && docker build -f Dockerfile.mcp -t signal-mcp .
	cd frontend && docker build -t signal-web .

build-all:
	docker-compose build

# ───────────────────────────────────────────
# PR Helper
# ───────────────────────────────────────────
pr:
	@read -p "Branch type (feat/fix/docs/chore/refactor): " type; \
	read -p "Short description (kebab-case): " desc; \
	branch="$$type/$$desc"; \
	git checkout -b $$branch; \
	git add .; \
	git commit -m "$$type: $$desc" || true; \
	git push -u origin $$branch; \
	gh pr create --title "$$type: $$desc" --body "$$desc" --base main

# ───────────────────────────────────────────
# Docs
# ───────────────────────────────────────────
docs:
	cd docs && npx mintlify dev

docs-deploy:
	cd docs && npx mintlify deploy
