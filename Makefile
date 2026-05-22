.PHONY: build test lint clean dev deps sqlc-gen mock-gen migrate-up migrate-down

# ─── 构建 ──────────────────────────────────────────────
build:
	go build -o bin/server ./cmd/server
	go build -o bin/sync ./cmd/sync
	go build -o bin/worker ./cmd/worker
	go build -o bin/migrator ./cmd/migrator
	go build -o bin/cli ./cmd/cli

build-server:
	go build -o bin/server ./cmd/server

build-all: build

# ─── 开发 ──────────────────────────────────────────────
dev:
	go run ./cmd/server

dev-hot:
	go run ./cmd/server -watch

# ─── 测试 ──────────────────────────────────────────────
test:
	go test -race -count=1 ./...

test-unit:
	go test -short -race -count=1 ./...

test-int:
	go test -tags=integration -race -count=1 ./...

test-cover:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# ─── 代码质量 ──────────────────────────────────────────
lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

# ─── 数据库 ────────────────────────────────────────────
migrate-up:
	migrate -path internal/repository/db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path internal/repository/db/migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir internal/repository/db/migrations -seq $$name

migrate-status:
	migrate -path internal/repository/db/migrations -database "$(DATABASE_URL)" version

sqlc-gen:
	sqlc generate

# ─── Mock 生成 ─────────────────────────────────────────
mock-gen:
	@echo "Generating mocks..."
	@go generate ./...

# ─── 依赖 ──────────────────────────────────────────────
deps:
	go mod tidy
	go mod download

deps-update:
	go get -u ./...
	go mod tidy

# ─── Docker ────────────────────────────────────────────
docker-up:
	docker compose -f infra/docker/docker-compose.dev.yml up -d

docker-down:
	docker compose -f infra/docker/docker-compose.dev.yml down

docker-logs:
	docker compose -f infra/docker/docker-compose.dev.yml logs -f

# ─── 清理 ──────────────────────────────────────────────
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

# ─── 全部 ──────────────────────────────────────────────
all: deps lint test build
