.PHONY: all build run dev test clean lint fmt swagger wire docker-build docker-run docker-dev docker-dev-d docker-dev-stop docker-dev-logs docker-dev-rebuild init migrate help

# 变量定义
APP_NAME := ticketdesk
BUILD_DIR := bin
MAIN_FILE := cmd/ticketdesk/main.go
CONFIG_FILE := configs/config.yaml

# Go 相关
GOCMD := go
GOBUILD := $(GOCMD) build
GORUN := $(GOCMD) run
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOIMPORTS := goimports

# 构建标志
LDFLAGS := -ldflags "-s -w"

# 默认目标
all: lint test build

# 帮助信息
help:
	@echo "TicketDesk Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make init          - 初始化项目（安装依赖和工具）"
	@echo "  make build         - 构建项目"
	@echo "  make run           - 运行项目"
	@echo "  make dev           - 开发模式（前后端热更新）"
	@echo "  make dev-backend   - 仅启动后端（热更新）"
	@echo "  make dev-frontend  - 仅启动前端（热更新）"
	@echo "  make test          - 运行测试"
	@echo "  make lint          - 代码静态检查"
	@echo "  make fmt           - 格式化代码"
	@echo "  make swagger       - 生成 API 文档"
	@echo "  make wire          - 生成依赖注入代码"
	@echo "  make migrate       - 运行数据库迁移"
	@echo "  make docker-build  - 构建 Docker 镜像"
	@echo "  make docker-run    - 运行 Docker 容器"
	@echo "  make docker-dev    - Docker 开发模式（前后端热更新）"
	@echo "  make docker-dev-d  - Docker 开发模式（后台运行）"
	@echo "  make docker-dev-stop - 停止 Docker 开发容器"
	@echo "  make docker-dev-logs - 查看 Docker 开发日志"
	@echo "  make clean         - 清理构建产物"
	@echo ""

# 初始化项目
init:
	@echo ">>> 安装后端依赖..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo ">>> 安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/google/wire/cmd/wire@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/air-verse/air@latest
	@echo ">>> 安装前端依赖..."
	cd web && npm install
	@echo ">>> 初始化完成"

# 构建
build:
	@echo ">>> 构建项目..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo ">>> 构建完成: $(BUILD_DIR)/$(APP_NAME)"

# 运行
run:
	@echo ">>> 运行项目..."
	$(GORUN) $(MAIN_FILE) -config $(CONFIG_FILE)

# 开发模式（热更新）- 同时启动前后端
dev:
	@echo ">>> 开发模式启动（前后端热更新）..."
	@bash scripts/dev.sh

# 仅启动后端（热更新）
dev-backend:
	@echo ">>> 后端开发模式启动（热更新）..."
	air -c .air.toml

# 仅启动前端（热更新）
dev-frontend:
	@echo ">>> 前端开发模式启动（热更新）..."
	cd web && npm run dev

# 测试
test:
	@echo ">>> 运行测试..."
	$(GOTEST) -v -race -cover ./...

# 测试覆盖率
test-coverage:
	@echo ">>> 生成测试覆盖率报告..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo ">>> 覆盖率报告: coverage.html"

# 代码检查
lint:
	@echo ">>> 代码静态检查..."
	golangci-lint run ./...

# 格式化代码
fmt:
	@echo ">>> 格式化代码..."
	$(GOFMT) -s -w .
	$(GOIMPORTS) -w .

# 生成 Swagger 文档
swagger:
	@echo ">>> 生成 API 文档..."
	swag init -g $(MAIN_FILE) -o docs/swagger
	@echo ">>> API 文档生成完成: docs/swagger"

# 生成 Wire 依赖注入代码
wire:
	@echo ">>> 生成依赖注入代码..."
	wire ./...

# 数据库迁移
migrate:
	@echo ">>> 运行数据库迁移..."
	$(GORUN) $(MAIN_FILE) -config $(CONFIG_FILE) -migrate

# Docker 构建
docker-build:
	@echo ">>> 构建 Docker 镜像..."
	docker build -t $(APP_NAME):latest -f deploy/docker/Dockerfile .

# Docker 运行（MySQL + Redis）
docker-run:
	@echo ">>> 运行 Docker 容器..."
	docker-compose -f deploy/docker-compose.yaml up -d

# Docker 停止
docker-stop:
	@echo ">>> 停止 Docker 容器..."
	docker-compose -f deploy/docker-compose.yaml down

# Docker 开发模式（前后端热更新）
docker-dev:
	@echo ">>> Docker 开发模式启动（前后端热更新）..."
	docker-compose -f deploy/docker-compose.dev.yaml up --build

# Docker 开发模式（后台运行）
docker-dev-d:
	@echo ">>> Docker 开发模式启动（后台运行）..."
	docker-compose -f deploy/docker-compose.dev.yaml up --build -d

# Docker 开发模式停止
docker-dev-stop:
	@echo ">>> 停止 Docker 开发容器..."
	docker-compose -f deploy/docker-compose.dev.yaml down

# Docker 开发模式日志
docker-dev-logs:
	@echo ">>> 查看 Docker 开发容器日志..."
	docker-compose -f deploy/docker-compose.dev.yaml logs -f

# Docker 开发模式重建
docker-dev-rebuild:
	@echo ">>> 重建 Docker 开发容器..."
	docker-compose -f deploy/docker-compose.dev.yaml up --build --force-recreate

# 清理
clean:
	@echo ">>> 清理构建产物..."
	rm -rf $(BUILD_DIR)
	rm -rf tmp/
	rm -f coverage.out coverage.html build-errors.log
	@echo ">>> 清理完成"
