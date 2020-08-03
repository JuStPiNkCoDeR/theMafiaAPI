include .env
###########################
# Basic params & commands #
###########################
SERVICE_PATH=$(PWD)
SERVICE_NAME=$(shell basename $(SERVICE_PATH))
LIB_PATH=$(SERVICE_PATH)/../../lib
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_CLEAN=$(GO_CMD) clean
GO_TEST=$(GO_CMD) test
GO_GET=$(GO_CMD) get
################
# Binary names #
################
BINARY_NAME=bin
BINARY_UNIX=$(BINARY_NAME)_unix
#########
# Tasks #
#########
.SILENT:
all: deps test run
build:
	echo "Building application..."
	$(GO_BUILD) -o $(BINARY_NAME) -v
test:
	echo "Testing appication..."
	$(GO_TEST) -v ./...
clean:
	echo "Cleaing..."
	$(GO_CLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GO_BUILD) -o $(BINARY_NAME) -v
	echo "Staring application..."
	./$(BINARY_NAME)
deps:
	echo "Install dependencies"
	$(GO_GET) -d -v ./...
# Cross-platform
build-linux:
	echo "Building for linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BINARY_NAME) -v
run-docker:
	echo "Running in docker..."
	cp -r $(LIB_PATH) ./
	docker-compose up -d $(SERVICE_NAME)