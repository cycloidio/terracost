IS_CI ?= 0

DOCKER_COMPOSE_CMD := docker-compose
GO_CMD := go
GO_TEST_CMD := $(GO_CMD) test ./...
GO_RUN_CMD := $(GO_CMD) run

.PHONY: test
test: db-up
	@$(GO_TEST_CMD)

.PHONY: db-up
db-up: # Start the DB server
ifeq ($(IS_CI), 0)
	@$(DOCKER_COMPOSE_CMD) up -d database
endif

.PHONY: down
down:
	@$(DOCKER_COMPOSE_CMD) down

.PHONY: db-migrate
db-migrate: db-up
	@$(GO_RUN_CMD) scripts/migrate.go

