BIN=bin
DAEMON_DIR=./cmd/daemon
CLIENT_DIR=./cmd/client

.PHONY: all windows linux win32 win64 proto build test clean run-server run-client lint lint-fix fmt vet

# Генерация protobuf кода
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/stats.proto

all: windows linux

# ---------- Windows ----------
windows: win32 win64

win32:
	@echo "==> Windows 32-bit"
	GOOS=windows GOARCH=386 go build -o $(BIN)/vsysmon32.exe $(DAEMON_DIR)
	GOOS=windows GOARCH=386 go build -o $(BIN)/client32.exe $(CLIENT_DIR)

win64:
	@echo "==> Windows 64-bit"
	GOOS=windows GOARCH=amd64 go build -o $(BIN)/vsysmon64.exe $(DAEMON_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN)/client64.exe $(CLIENT_DIR)

# ---------- Linux ----------
linux:
	@echo "==> Linux amd64"
	GOOS=linux GOARCH=amd64 go build -o $(BIN)/vsysmon $(DAEMON_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN)/client $(CLIENT_DIR)

# Запуск тестов
test:
	go test -v ./...

# Запуск тестов с покрытием
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Запуск тестов с race detector
test-race:
	go test -v -race ./...

# Линтинг
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Линтинг с автоисправлением
lint-fix:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Форматирование кода
fmt:
	go fmt ./...
	@if command -v gofumpt >/dev/null 2>&1; then \
		gofumpt -w .; \
	fi

# Проверка кода с помощью go vet
vet:
	go vet ./...

# Проверка всех (lint + vet + test)
check: lint vet test

# Очистка
clean:
	rm -rf bin/ coverage.out coverage.html

# Запуск сервера
run-server:
	go run ./cmd/daemon

# Запуск клиента
run-client:
	go run ./cmd/client

# Docker команды
docker-build:
	docker build -t vsysmon:latest .

docker-run:
	docker run -d --name vsysmon --privileged -p 50051:50051 \
		-v $(PWD)/bin/config.json:/etc/vsysmon/config.json:ro \
		vsysmon:latest

docker-stop:
	docker stop vsysmon
	docker rm vsysmon

docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-logs:
	docker logs -f vsysmon-daemon