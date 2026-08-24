.PHONY: help deps deps-backend deps-agent deps-frontend infra-up infra-down infra-logs migrate-up api worker agent frontend test test-backend test-agent lint build build-backend build-agent build-frontend clean

help:
	@echo "Caelus Cloud - Development Tasks"
	@echo ""
	@echo "Dependency Management:"
	@echo "  make deps           - Download all dependencies (Backend, Agent, Frontend)"
	@echo "  make deps-backend   - Download Go modules for backend"
	@echo "  make deps-agent     - Download Go modules for agent"
	@echo "  make deps-frontend  - Install pnpm packages for frontend"
	@echo ""
	@echo "Infrastructure (Docker Compose):"
	@echo "  make infra-up       - Start PostgreSQL, Redis, MinIO in background"
	@echo "  make infra-down     - Stop all infrastructure containers"
	@echo "  make infra-logs     - Follow infrastructure service logs"
	@echo ""
	@echo "Database Migrations:"
	@echo "  make migrate-up     - Run pending database migrations (Up)"
	@echo ""
	@echo "Application Services:"
	@echo "  make api            - Run Backend REST API Server (port 8080)"
	@echo "  make worker         - Run Background Worker & Task Scheduler"
	@echo "  make agent          - Run Host Telemetry Agent Daemon"
	@echo "  make frontend       - Run Frontend Next.js Dev Server (port 3000)"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test           - Run unit tests for backend & agent"
	@echo "  make test-backend   - Run unit tests for backend"
	@echo "  make test-agent     - Run unit tests for agent"
	@echo "  make lint           - Lint frontend TypeScript code"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Compile backend, agent, and build frontend"
	@echo "  make clean          - Remove build artifacts and temporary files"

# ------------------------------------------------------------------------------
# Dependencies
# ------------------------------------------------------------------------------
deps: deps-backend deps-agent deps-frontend
	@echo "==> Seluruh dependensi (Backend, Agent, Frontend) berhasil disiapkan!"

deps-backend:
	@echo "==> Downloading backend Go dependencies..."
	@cd backend && go mod download && go mod verify
	@echo "==> Backend Go dependencies: OK (terverifikasi)"

deps-agent:
	@echo "==> Downloading agent Go dependencies..."
	@cd agent && go mod download && go mod verify
	@echo "==> Agent Go dependencies: OK (terverifikasi)"

deps-frontend:
	@echo "==> Installing frontend dependencies via pnpm..."
	@cd frontend && pnpm install
	@echo "==> Frontend dependencies: OK"

# ------------------------------------------------------------------------------
# Infrastructure
# ------------------------------------------------------------------------------
infra-up:
	@echo "==> Starting local infrastructure (PostgreSQL, Redis, MinIO)..."
	@docker compose up -d postgres redis minio

infra-down:
	@echo "==> Stopping local infrastructure..."
	@docker compose down

infra-logs:
	@docker compose logs -f postgres redis minio

# ------------------------------------------------------------------------------
# Database Migrations
# ------------------------------------------------------------------------------
migrate-up:
	@echo "==> Running database migrations up..."
	@cd backend && go run cmd/migrate/main.go -direction=up

# ------------------------------------------------------------------------------
# Application Services
# ------------------------------------------------------------------------------
api:
	@echo "==> Starting Caelus API server on http://localhost:8080..."
	@cd backend && go run cmd/api/main.go

worker:
	@echo "==> Starting Caelus Worker & Task Scheduler..."
	@cd backend && go run cmd/worker/main.go

agent:
	@echo "==> Starting Caelus Host Telemetry Agent..."
	@cd agent && go run cmd/main.go

frontend:
	@echo "==> Starting Caelus Frontend on http://localhost:3000..."
	@cd frontend && pnpm dev

# ------------------------------------------------------------------------------
# Testing & Quality
# ------------------------------------------------------------------------------
test: test-backend test-agent

test-backend:
	@echo "==> Running backend tests..."
	@cd backend && go test -v ./tests/...

test-agent:
	@echo "==> Running agent tests..."
	@cd agent && go test -v ./tests/...

lint:
	@echo "==> Linting frontend..."
	@cd frontend && pnpm lint

# ------------------------------------------------------------------------------
# Build & Clean
# ------------------------------------------------------------------------------
build: build-backend build-agent build-frontend

build-backend:
	@echo "==> Building backend binaries..."
	@mkdir -p backend/bin
	@cd backend && go build -o bin/caelus-api cmd/api/main.go
	@cd backend && go build -o bin/caelus-worker cmd/worker/main.go
	@cd backend && go build -o bin/caelus-migrate cmd/migrate/main.go

build-agent:
	@echo "==> Building agent binary..."
	@mkdir -p agent/bin
	@cd agent && go build -o bin/caelus-agent cmd/main.go

build-frontend:
	@echo "==> Building frontend production bundle..."
	@cd frontend && pnpm build

# ------------------------------------------------------------------------------
# Docker Images Build
# ------------------------------------------------------------------------------
docker-build: docker-build-api docker-build-worker docker-build-agent docker-build-ui
	@echo "==> Seluruh Docker Images Caelus Cloud berhasil dibangun!"

docker-build-api:
	@echo "==> Building Docker Image: caelus-api:latest..."
	@docker build -t caelus-api:latest -f deploy/docker/Dockerfile.api backend

docker-build-worker:
	@echo "==> Building Docker Image: caelus-worker:latest..."
	@docker build -t caelus-worker:latest -f deploy/docker/Dockerfile.worker backend

docker-build-agent:
	@echo "==> Building Docker Image: caelus-agent:latest..."
	@docker build -t caelus-agent:latest -f deploy/docker/Dockerfile.agent agent

docker-build-ui:
	@echo "==> Building Docker Image: caelus-frontend:latest..."
	@docker build -t caelus-frontend:latest -f deploy/docker/Dockerfile.ui frontend

clean:
	@echo "==> Cleaning build artifacts..."
	@rm -rf backend/bin agent/bin frontend/.next frontend/out
