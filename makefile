BIN=bin
BUILD_DIR=bin
CLIENT_DIR=./client

.PHONY: all windows linux win32 win64 clean

all: windows linux

# ---------- Windows ----------
windows: win32 win64

win32:
	@echo "==> Windows 32-bit"
	GOOS=windows GOARCH=386 go build -o $(BIN)/vsysmon32.exe
	GOOS=windows GOARCH=386 go build -o $(BIN)/client32.exe $(CLIENT_DIR)

win64:
	@echo "==> Windows 64-bit"
	GOOS=windows GOARCH=amd64 go build -o $(BIN)/vsysmon64.exe
	GOOS=windows GOARCH=amd64 go build -o $(BIN)/client64.exe $(CLIENT_DIR)

# ---------- Linux ----------
linux:
	@echo "==> Linux amd64"
	GOOS=linux GOARCH=amd64 go build -o $(BIN)/vsysmon
	GOOS=linux GOARCH=amd64 go build -o $(BIN)/client $(CLIENT_DIR)

clean:
	@echo "==> clean"
	rm -rf $(BUILD_DIR)
