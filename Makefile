.PHONY: help build run test clean docker

# 默认目标
.DEFAULT_GOAL := help

# 项目信息
APP_NAME := go-uni-pay
VERSION := 1.0.0
BUILD_DIR := bin
CONFIG_FILE := configs/config.yaml

# Go 相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# 构建目标
BINARY_NAME := $(APP_NAME)
BINARY_UNIX := $(BINARY_NAME)_unix

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## 安装依赖
	$(GOMOD) download
	$(GOMOD) tidy

build: ## 编译项目
	@echo "开始编译..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v cmd/server/main.go
	@echo "编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

run: ## 运行项目
	@echo "启动服务..."
	$(GOCMD) run cmd/server/main.go $(CONFIG_FILE)

dev: ## 开发模式运行（使用 air 热重载，需先安装 air）
	@which air > /dev/null || (echo "请先安装 air: go install github.com/cosmtrek/air@latest" && exit 1)
	air

test: ## 运行测试
	$(GOTEST) -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

clean: ## 清理编译文件
	@echo "清理编译文件..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "清理完成"

db-init: ## 初始化数据库
	@echo "初始化数据库..."
	@mysql -u root -p < scripts/init.sql
	@echo "数据库初始化完成"

docker-build: ## 构建 Docker 镜像
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

docker-run: ## 运行 Docker 容器
	docker-compose up -d

docker-stop: ## 停止 Docker 容器
	docker-compose down

docker-logs: ## 查看 Docker 容器日志
	docker-compose logs -f

lint: ## 代码检查
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint" && exit 1)
	golangci-lint run

fmt: ## 格式化代码
	$(GOCMD) fmt ./...
	goimports -w .

install: build ## 安装到系统
	@echo "安装到系统..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "安装完成"

uninstall: ## 从系统卸载
	@echo "从系统卸载..."
	rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "卸载完成"

# 跨平台编译
build-linux: ## 编译 Linux 版本
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) -v cmd/server/main.go

build-windows: ## 编译 Windows 版本
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME).exe -v cmd/server/main.go

build-mac: ## 编译 MacOS 版本
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_darwin -v cmd/server/main.go

build-all: build-linux build-windows build-mac ## 编译所有平台版本
