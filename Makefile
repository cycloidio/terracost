IS_CI ?= 0

DOCKER_COMPOSE_CMD := docker-compose
GO_CMD := go
GO_TEST_CMD := $(GO_CMD) test
GO_RUN_CMD := $(GO_CMD) run
GOLINT_CMD := $(GO_RUN_CMD) golang.org/x/lint/golint
GOIMPORTS_CMD := $(GO_RUN_CMD) golang.org/x/tools/cmd/goimports

.PHONY: ci
ci: lint test

.PHONY: test
test: db-up
	@$(GO_TEST_CMD) ./...

.PHONY: lint
lint:
	@$(GOLINT_CMD) -set_exit_status ./... && test -z "`$(GO_CMD) list -f {{.Dir}} ./... | xargs $(GOIMPORTS_CMD) -l | tee /dev/stderr`"

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

.PHONY: generate
generate:
	@rm -rf ./mock/
	@$(GO_CMD) generate ./...
	@$(GOIMPORTS_CMD) -w ./mock

.PHOHY: goimports
goimports:
	@$(GOIMPORTS_CMD) -w ./
