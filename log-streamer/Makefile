# Variables
FRONTEND_DIR = frontend
STATIC_DIR = static
BUILD_DIR = $(FRONTEND_DIR)/build
GO_BINARY = log-streamer

# Targets
all: build-frontend move-build-files build-go run

build-frontend:
	@echo "Building the React frontend..."
	cd $(FRONTEND_DIR) && npm install && npm run build

move-build-files:
	@echo "Moving build files to the static directory..."
	mkdir -p $(STATIC_DIR)
	rm -rf $(STATIC_DIR)/*
	mv $(BUILD_DIR)/* $(STATIC_DIR)/

build-go:
	@echo "Building the Go binary..."
	go build -o $(GO_BINARY) -buildvcs=false

run:
	@echo "Starting the application..."
	./$(GO_BINARY)

clean:
	@echo "Cleaning up..."
	rm -rf $(STATIC_DIR) $(GO_BINARY)
	cd $(FRONTEND_DIR) && rm -rf node_modules $(BUILD_DIR)

.PHONY: all build-frontend move-build-files build-go run clean
