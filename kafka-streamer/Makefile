# Variables
FRONTEND_DIR = kafka-frontend
STATIC_DIR = static
BUILD_DIR = $(FRONTEND_DIR)/build
GO_BINARY = kafka-server/kafka-server

# Targets
all: build-frontend move-build-files build-go run

build-frontend:
	@echo "Building the React frontend..."
	cd $(FRONTEND_DIR) && npm install && npm install moment react-json-tree && npm run build

move-build-files:
	@echo "Moving build files to the static directory..."
	mkdir -p $(STATIC_DIR)
	rm -rf $(STATIC_DIR)/*
	mv $(BUILD_DIR)/* $(STATIC_DIR)/

build-go:
	@echo "Building the Go binary..."
	cd kafka-server && go get && go build -o kafka-server -buildvcs=false

run:
	@echo "Starting the application..."
	cd kafka-server && ./kafka-server & \
	cd kafka-frontend && npm start

clean:
	@echo "Cleaning up..."
	rm -rf $(STATIC_DIR) $(GO_BINARY)
	cd $(FRONTEND_DIR) && rm -rf node_modules $(BUILD_DIR)

.PHONY: all build-frontend move-build-files build-go run clean
