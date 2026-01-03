.PHONY: run dev build clean up down swag test lint docker-build docker-push help

# Variables
APP_NAME := oracle-backend
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DOCKER_REGISTRY ?= ghcr.io/ifa-labs
DOCKER_IMAGE := $(DOCKER_REGISTRY)/oracle_engine

# Go variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
LDFLAGS := -ldflags "-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

## help: Show this help message
help:
	@echo "Oracle Engine - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## run: Run development environment with docker compose
run:
	docker compose -f docker-compose.dev.yml up --build -d

## dev: Run development environment in foreground
dev:
	docker compose -f docker-compose.dev.yml up --build

## build: Build docker images
build:
	docker compose build

## build-local: Build binary locally
build-local:
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME) cmd/oracle/main.go

## swag: Generate Swagger documentation
swag:
	swag init -g internal/server/api/api.go --output docs

## swag-clean: Remove generated Swagger docs
swag-clean:
	rm -rf docs/

## test: Run tests
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## test-short: Run tests without race detector
test-short:
	$(GOTEST) -v -coverprofile=coverage.out ./...

## coverage: Run tests and show coverage report
coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linter
lint:
	golangci-lint run --timeout 5m

## lint-fix: Run linter and fix issues
lint-fix:
	golangci-lint run --fix --timeout 5m

## clean: Clean up containers, volumes, and build artifacts
clean:
	docker compose -f docker-compose.dev.yml down -v
	docker compose -f docker-compose.prod.yml down -v
	rm -rf bin/ tmp/ coverage.out coverage.html

## up: Start production environment
up:
	docker compose -f docker-compose.prod.yml up -d

## down: Stop all containers
down:
	docker compose -f docker-compose.dev.yml down
	docker compose -f docker-compose.prod.yml down

## logs: Show logs from oracle service
logs:
	docker compose -f docker-compose.dev.yml logs -f oracle

## logs-prod: Show logs from production oracle service
logs-prod:
	docker compose -f docker-compose.prod.yml logs -f oracle

## docker-build: Build production Docker image
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-t $(DOCKER_IMAGE):latest \
		.

## docker-push: Push Docker image to registry
docker-push:
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

## mod-tidy: Tidy go modules
mod-tidy:
	$(GOMOD) tidy
	$(GOMOD) verify

## mod-update: Update all dependencies
mod-update:
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

## security-scan: Run security scan with trivy
security-scan:
	trivy fs --severity HIGH,CRITICAL .

## db-migrate: Run database migrations (placeholder)
db-migrate:
	@echo "Database migration not implemented yet"

## db-shell: Connect to database shell
db-shell:
	docker compose exec timescale psql -U $${POSTGRES_USER} -d $${POSTGRES_DB}

## redis-shell: Connect to redis shell
redis-shell:
	docker compose exec redis redis-cli

## db-backup: Create database backup
db-backup:
	./scripts/deploy/backup-db.sh

# ==============================================================================
# Deployment Commands
# ==============================================================================

## deploy-staging: Deploy to staging server
deploy-staging:
	@if [ -z "$(DEPLOY_HOST)" ]; then \
		echo "Error: DEPLOY_HOST is not set"; \
		echo "Usage: make deploy-staging DEPLOY_HOST=staging.example.com"; \
		exit 1; \
	fi
	DEPLOY_HOST=$(DEPLOY_HOST) ./scripts/deploy/deploy.sh

## deploy-prod: Deploy to production server
deploy-prod:
	@if [ -z "$(DEPLOY_HOST)" ]; then \
		echo "Error: DEPLOY_HOST is not set"; \
		echo "Usage: make deploy-prod DEPLOY_HOST=api.example.com"; \
		exit 1; \
	fi
	@echo "⚠️  Deploying to PRODUCTION. Are you sure? [y/N]" && read ans && [ $${ans:-N} = y ]
	DEPLOY_HOST=$(DEPLOY_HOST) COMPOSE_FILE=docker-compose.prod.yml ./scripts/deploy/deploy.sh

## rollback: Rollback to a specific version
rollback:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is not set"; \
		echo "Usage: make rollback VERSION=v1.2.3 DEPLOY_HOST=api.example.com"; \
		exit 1; \
	fi
	@if [ -z "$(DEPLOY_HOST)" ]; then \
		echo "Error: DEPLOY_HOST is not set"; \
		exit 1; \
	fi
	DEPLOY_HOST=$(DEPLOY_HOST) ./scripts/deploy/rollback.sh $(VERSION)

## setup-server: Run server setup script (requires root)
setup-server:
	@echo "This command should be run on the target server"
	@echo "Copy scripts/deploy/setup-server.sh to the server and run as root"

## release: Create a new release tag
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is not set"; \
		echo "Usage: make release VERSION=v1.2.3"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "Release $(VERSION) created and pushed!"

