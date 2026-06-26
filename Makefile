VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME = $(shell date -u +%Y-%m-%dT%H:%M:%S)
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS = -X godis/version.Version=$(VERSION) \
          -X godis/version.BuildTime=$(BUILD_TIME) \
          -X godis/version.GitCommit=$(GIT_COMMIT)

OUTPUT_DIR = build

.PHONY: all build clean test test-integration build-all docker-build

# Docker 构建
docker-build:
	docker build -t godis:latest .

# Docker 运行
docker-run:
	docker run -d -p 6379:6379 -v godis-data:/data -v godis-logs:/logs --name godis godis:latest

# 默认构建（当前平台）
build:
	go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/godis .

# 本地构建并运行
run: build
	./$(OUTPUT_DIR)/godis --config ./etc/godis.yaml

# 交叉编译所有平台
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux-amd64:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/godis-linux-amd64 .

build-linux-arm64:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/godis-linux-arm64 .

build-darwin-amd64:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/godis-darwin-amd64 .

build-darwin-arm64:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/godis-darwin-arm64 .

build-windows-amd64:
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/godis-windows-amd64.exe .

# 单元测试
test:
	go test ./...

# 集成测试
test-integration:
	go test -v ./integration/ -count=1

# 清理构建产物
clean:
	rm -rf $(OUTPUT_DIR)
