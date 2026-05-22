# 开发环境搭建指南

本文档指导你从零搭建 NextM 的本地开发环境。

---

## 目录

1. [前置依赖](#1-前置依赖)
2. [Go 后端](#2-go-后端)
3. [前端 Web](#3-前端-web)
4. [桌面端（Tauri）](#4-桌面端tauri)
5. [移动端（Flutter）](#5-移动端flutter)
6. [浏览器扩展](#6-浏览器扩展)
7. [基础设施服务](#7-基础设施服务)
8. [Makefile 常用命令](#8-makefile-常用命令)
9. [常见问题](#9-常见问题)

---

## 1. 前置依赖

| 工具 | 最低版本 | 用途 |
|------|---------|------|
| Git | 2.40+ | 版本控制 |
| Go | 1.24+ | 后端编译 |
| Node.js | 20 LTS+ | Web 前端 / 浏览器扩展 |
| pnpm | 9+ | JavaScript 包管理（推荐） |
| Docker | 24+ | 基础设施服务（PostgreSQL、NATS 等） |
| Rust | 1.80+ | Tauri 桌面端（可选） |
| Flutter | 3.24+ | 移动端（可选） |

### 1.1 安装 Go

```bash
# 推荐使用官方安装包或 gvm
# https://go.dev/dl/
go version  # 确认版本 >= 1.24
```

配置 Go 环境变量（`~/.bashrc` 或 `~/.zshrc`）：

```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### 1.2 安装 Node.js

```bash
# 推荐使用 nvm 管理版本
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash
nvm install 22
nvm use 22

# 安装 pnpm
corepack enable
corepack prepare pnpm@latest --activate
pnpm --version  # 确认安装成功
```

### 1.3 安装 Docker

```bash
# Docker Desktop for Windows / macOS 或 Docker Engine for Linux
# https://docs.docker.com/get-docker/
docker --version    # 确认安装成功
docker compose version  # 确认 docker compose 可用
```

---

## 2. Go 后端

### 2.1 安装工具链

```bash
# 安装 golangci-lint（代码检查）
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装 sqlc（SQL 代码生成）
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# 安装 golang-migrate（数据库迁移）
go install -tags 'sqlite3 postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 安装 mockgen（mock 生成）
go install go.uber.org/mock/mockgen@latest

# 安装 cobra CLI（可选，用于脚手架）
go install github.com/spf13/cobra-cli@latest
```

### 2.2 编译

```bash
# 编译所有二进制
go build ./...

# 编译特定服务
go build -o bin/server ./cmd/server
go build -o bin/sync ./cmd/sync
go build -o bin/worker ./cmd/worker

# 安装（到 $GOPATH/bin）
go install ./...
```

### 2.3 配置

创建本地配置文件：

```bash
# 从模板复制
cp .env.example .env

# 默认配置即可本地开发
# 如需覆盖，编辑 config/config.yaml 或设置环境变量
```

### 2.4 数据库

```bash
# 启动基础设施
docker compose -f infra/docker/docker-compose.dev.yml up -d

# 运行迁移
make migrate-up

# 回滚迁移
make migrate-down

# 查看迁移状态
make migrate-status
```

---

## 3. 前端 Web

```bash
# 进入 web 目录
cd frontend/web

# 安装依赖
pnpm install

# 启动开发服务器（默认 http://localhost:5173）
pnpm dev

# 类型检查
pnpm typecheck

# Lint
pnpm lint
```

环境变量配置 `frontend/web/.env.local`：

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
```

---

## 4. 桌面端（Tauri）

需要 Rust 工具链：

```bash
# 安装 Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# 安装 Tauri CLI
cargo install tauri-cli --version "^2"

# 进入桌面端目录
cd frontend/desktop

# 安装前端依赖
pnpm install

# 启动 Tauri 开发模式
pnpm tauri dev
```

---

## 5. 移动端（Flutter）

```bash
# 确保 Flutter 已安装并包含在 PATH 中
flutter doctor  # 确认所有检查项通过

# 进入移动端目录
cd frontend/mobile

# 获取依赖
flutter pub get

# 运行（需连接设备或模拟器）
flutter run
```

---

## 6. 浏览器扩展

```bash
cd frontend/extension

# 安装依赖
pnpm install

# 构建开发版本（加载到浏览器即可使用）
pnpm dev

# 构建生产版本
pnpm build
```

---

## 7. 基础设施服务

使用 Docker Compose 启动所有依赖服务：

```bash
# 启动所有服务
docker compose -f infra/docker/docker-compose.dev.yml up -d

# 查看服务状态
docker compose -f infra/docker/docker-compose.dev.yml ps

# 停止服务
docker compose -f infra/docker/docker-compose.dev.yml down
```

### 服务列表

| 服务 | 端口 | 说明 |
|------|------|------|
| PostgreSQL | 5432 | 云端数据库 |
| NATS | 4222 | 消息总线 |
| Redis | 6379 | Asynq 任务队列 |
| Meilisearch | 7700 | 全文搜索引擎 |
| LiteLLM | 4000 | AI 模型网关 |

---

## 8. Makefile 常用命令

```bash
make deps          # 安装所有依赖
make dev           # 启动所有开发服务
make build         # 构建所有可执行文件
make test          # 运行所有测试
make test-unit     # 运行单元测试
make test-int      # 运行集成测试（需 Docker）
make lint          # 运行所有 linter
make fmt           # 格式化代码
make migrate-up    # 执行数据库迁移
make migrate-down  # 回滚数据库迁移
make sqlc-gen      # 重新生成 SQLc Go 代码
make clean         # 清理构建产物
```

---

## 9. 常见问题

### Q: CGo 编译失败（go-sqlite3）

确保已安装 GCC/MinGW（Windows）或 build-essential（Linux）：

```bash
# Ubuntu/Debian
sudo apt install build-essential

# macOS
xcode-select --install

# Windows — 安装 MinGW-w64 并添加到 PATH
```

### Q: NATS 连接被拒绝

确保 Docker 服务正在运行：

```bash
docker ps | grep nats
```

### Q: Go 版本过低

```bash
go version
# 如果低于 1.24，请使用 gvm 升级或从 https://go.dev/dl 下载
```
