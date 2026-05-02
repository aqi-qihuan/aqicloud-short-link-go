.PHONY: all build test lint clean fmt vet install-hooks run-% docker-build docker-up docker-down help

SERVICES := gateway account link shop data ai

all: lint test build

## ==================== 构建 ====================

build:
	@mkdir -p bin
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		CGO_ENABLED=0 go build -o bin/$$svc ./cmd/$$svc/; \
	done
	@echo "✅ All services built to bin/"

build-%:
	@echo "Building $*..."
	@mkdir -p bin
	@CGO_ENABLED=0 go build -o bin/$* ./cmd/$*/

run-%:
	go run ./cmd/$*/main.go

## ==================== 代码质量 ====================

fmt:
	@echo "📝 格式化代码..."
	@gofmt -w .
	@echo "✅ 格式化完成"

vet:
	@echo "🔬 运行 go vet..."
	@go vet ./...
	@echo "✅ go vet 通过"

lint:
	@echo "🔧 运行 golangci-lint..."
	@golangci-lint run ./...
	@echo "✅ Lint 通过"

lint-install:
	@echo "安装 golangci-lint..."
	@brew install golangci-lint 2>/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ golangci-lint 安装完成"

## ==================== 测试 ====================

test:
	@echo "🧪 运行测试..."
	@go test ./... -v -count=1 -timeout=5m
	@echo "✅ 测试通过"

test-short:
	@go test ./... -short -count=1

test-cover:
	@echo "🧪 运行测试（含覆盖率）..."
	@go test ./... -coverprofile=coverage.out -count=1
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告: coverage.html"

## ==================== Git ====================

install-hooks:
	@echo "🔗 安装 Git Hooks..."
	@git config core.hooksPath scripts/githooks
	@echo "✅ Git Hooks 已安装（pre-commit + commit-msg）"

## ==================== Docker ====================

docker-build:
	@echo "🐳 构建 Docker 镜像..."
	@docker-compose build

docker-up:
	@echo "🐳 启动服务..."
	@docker-compose up -d

docker-down:
	@echo "🐳 停止服务..."
	@docker-compose down

## ==================== 工具 ====================

clean:
	@rm -rf bin/ coverage.out coverage.html
	@echo "✅ 清理完成"

mod-tidy:
	@go mod tidy
	@echo "✅ go mod tidy 完成"

help:
	@echo ""
	@echo "AqiCloud Short-Link Go — 可用命令:"
	@echo ""
	@echo "  构建:"
	@echo "    make build          构建所有服务到 bin/"
	@echo "    make build-link     构建单个服务"
	@echo "    make run-link       运行单个服务"
	@echo ""
	@echo "  代码质量:"
	@echo "    make fmt            格式化代码"
	@echo "    make vet            运行 go vet"
	@echo "    make lint           运行 golangci-lint"
	@echo "    make lint-install   安装 golangci-lint"
	@echo ""
	@echo "  测试:"
	@echo "    make test           运行所有测试"
	@echo "    make test-short     运行短测试"
	@echo "    make test-cover     运行测试并生成覆盖率报告"
	@echo ""
	@echo "  Git:"
	@echo "    make install-hooks  安装 Git Hooks"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-build   构建 Docker 镜像"
	@echo "    make docker-up      启动所有服务"
	@echo "    make docker-down    停止所有服务"
	@echo ""
	@echo "  工具:"
	@echo "    make clean          清理构建产物"
	@echo "    make mod-tidy       整理 Go 依赖"
	@echo ""
