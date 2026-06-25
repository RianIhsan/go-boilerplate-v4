.PHONY: run build test mock lint migrate-up migrate-down tidy vet vulncheck check install-hooks

# ── Run ───────────────────────────────────────────────────
run:
	go run ./cmd/api/main.go

build:
	go build -o bin/api ./cmd/api/main.go

# ── Test ─────────────────────────────────────────────────
test:
	go test ./... -v -cover

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# ── Mock generation ──────────────────────────────────────
mock:
	go generate ./internal/domain/auth/repository/...
	go generate ./internal/domain/auth/usecase/...
	go generate ./internal/domain/todo/repository/...
	go generate ./internal/domain/todo/usecase/...
	go generate ./internal/domain/filemeta/usecase/...

# ── Database migrations ───────────────────────────────────
migrate-up:
	migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down:
	migrate -path ./migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

# ── Deps ─────────────────────────────────────────────────
tidy:
	go mod tidy

lint:
	golangci-lint run ./...

vet:
	go vet ./...

vulncheck:
	govulncheck ./...

# ── Pre-commit / pre-push gate ───────────────────────────
# Mirrors the CI workflow (.github/workflows/ci.yml) exactly, so a clean
# `make check` locally means CI will be green too.
check:
	go build ./...
	go vet ./...
	go test ./... -race -cover
	golangci-lint run ./...
	govulncheck ./...
	@echo "All checks passed."

install-hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks path set to .githooks/ — pre-commit checks are now active for this clone."
