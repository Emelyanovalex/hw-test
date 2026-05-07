BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

# Postgres
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB ?= calendar
POSTGRES_PORT ?= 5435
POSTGRES_HOST ?= localhost
POSTGRES_CONTAINER := postgres-calendar
POSTGRES_DSN ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

# RabbitMQ (used in later homework)
RABBIT_USER ?= rabbit
RABBIT_PASSWORD ?= password
RABBIT_PORT ?= 5672
RABBIT_PORT_UI ?= 15672
RABBIT_CONTAINER := rabbitmq

run-postgres:
	docker run -d --name $(POSTGRES_CONTAINER) \
	-e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	-e POSTGRES_DB=$(POSTGRES_DB) \
	-p $(POSTGRES_PORT):5432 \
	-v postgres-data:/var/lib/postgresql/data \
	postgres:16

run-rabbit:
	docker run -d --name $(RABBIT_CONTAINER) \
	-p $(RABBIT_PORT):5672 \
	-p $(RABBIT_PORT_UI):15672 \
	-e RABBITMQ_DEFAULT_USER=$(RABBIT_USER) \
	-e RABBITMQ_DEFAULT_PASS=$(RABBIT_PASSWORD) \
	rabbitmq:management

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race -count 100 ./internal/...

GOLANGCI_LINT_VERSION ?= v1.63.4

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

lint: install-lint-deps
	golangci-lint run ./...

install-goose:
	(which goose > /dev/null) || go install github.com/pressly/goose/v3/cmd/goose@latest

migrate: install-goose
	goose -dir ./migrations postgres "$(POSTGRES_DSN)" up

.PHONY: build run build-img run-img version test lint install-lint-deps install-goose migrate run-postgres run-rabbit
