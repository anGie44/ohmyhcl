NAME := tfrefactor
SRC_DIR := ./tfrefactor

.DEFAULT_GOAL := build

.PHONY: deps
deps:
	cd $(SRC_DIR) && go mod download

.PHONY: build
build: deps
	cd $(SRC_DIR) && go build -o ~/go/bin/$(NAME)

.PHONY: install
install: deps
	cd $(SRC_DIR) && go install

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test: build
	go test ./$(SRC_DIR)/...

.PHONY: check
check: lint test