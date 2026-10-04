BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

#Postrges
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB ?= backend
POSTGRES_PORT ?= 5435
POSTGRES_CONTAINER := postgres-calendar
POSTGRES_SSLMODE ?= disable

#Migrations
MIGRATIONS_DIR ?= ./migrations
GOOSE_IMAGE ?= golang:1.26
GOOSE_VERSION ?= latest
GOOSE_DSN := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@127.0.0.1:5432/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)
GOLANGCI_LINT := $(shell go env GOPATH)/bin/golangci-lint
GOLANGCI_LINT_DIR := $(dir $(GOLANGCI_LINT))
GOLANGCI_LINT_VERSION ?= v2.13.0

#Rabbit
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

migrate:
	docker run --rm \
		--network container:$(POSTGRES_CONTAINER) \
		-v $(CURDIR)/$(MIGRATIONS_DIR):/migrations:ro \
		$(GOOSE_IMAGE) \
		sh -c 'go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) && goose -dir /migrations postgres "$(GOOSE_DSN)" up'

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
	GOCACHE=/tmp/go-build-cache go test -race ./internal/...

install-lint-deps:
	(test -x $(GOLANGCI_LINT)) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOLANGCI_LINT_DIR) $(GOLANGCI_LINT_VERSION)

lint: install-lint-deps
	GOCACHE=/tmp/go-build-cache $(GOLANGCI_LINT) run ./...

.PHONY: build run migrate build-img run-img version test lint
