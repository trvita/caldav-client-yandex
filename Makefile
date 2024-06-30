# Variables
BINARY_NAME := myclient
BUILD_DIR := build
CMD_DIR := cmd/caldav-client

all: build

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

test:
	go test $(CMD_DIR)

clean:
	rm -rf $(BUILD_DIR)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: all build test clean run
