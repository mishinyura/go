.PHONY: build run stop tidy fmt vet

ENV_FILE ?= .env

stop:
	@-pkill -f "go run.*cmd/server" 2>/dev/null || true
	@-pkill -f "bin/server" 2>/dev/null || true

build:
	go build -o bin/server ./cmd/server

run: stop
	@if [ -f $(ENV_FILE) ]; then \
		export $$(grep -v '^\s*#' $(ENV_FILE) | grep -v '^\s*$$' | xargs); \
	fi; \
	go run ./cmd/server

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...


