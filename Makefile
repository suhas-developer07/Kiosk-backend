APP_NAME := kiosk-backend
MAIN := ./src/main.go
BUILD_DIR := bin
PKG := ./...

.PHONY: all build run clean test fmt tidy deps

## -------------------------
## Default target
## -------------------------
all: build

## -------------------------
## Build the app
## -------------------------
build:
	@echo "🚧 Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)
	@echo "✅ Build completed: $(BUILD_DIR)/$(APP_NAME)"

## -------------------------
## Run the app
## -------------------------
run:
	@echo "🏃 Running $(APP_NAME)..."
	@go run $(MAIN)

## -------------------------
## Run tests
## -------------------------
test:
	@echo "🧪 Running tests..."
	@go test $(PKG) -v

## -------------------------
## Format code
## -------------------------
fmt:
	@echo "🧹 Formatting source files..."
	@go fmt $(PKG)

## -------------------------
## Tidy modules
## -------------------------
tidy:
	@echo "🔧 Tidying go.mod..."
	@go mod tidy

## -------------------------
## Download dependencies
## -------------------------
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download

## -------------------------
## Clean build files
## -------------------------
clean:
	@echo "🗑 Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "✔ Clean completed"
