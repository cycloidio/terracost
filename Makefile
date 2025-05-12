IS_CI ?= 0
TOOL_BIN := $(PWD)/bin

DOCKER_COMPOSE_CMD := docker-compose
GO_CMD := go
GO_TEST_CMD := $(GO_CMD) test
GO_RUN_CMD := $(GO_CMD) run

MYSQL_USER := root
MYSQL_PASS := terracost
MYSQL_DB := terracost_test
MYSQL_DUMP ?= mysql/testdata/2023-02-23-pricing.sql.gz

BIN_DIR := $(GOPATH)/bin

GOLINT := $(TOOL_BIN)/golangci-lint
GOIMPORTS := $(BIN_DIR)/goimports
ENUMER := $(BIN_DIR)/enumer
MOCKGEN := $(BIN_DIR)/mockgen

# The ci rule is voluntarily simple because
# the CI requires to run the command separately
# and not all at once, otherwise it wouldn't work
.PHONY: ci
ci: lint
	@$(GO_TEST_CMD) ./...

.PHONY: test
test: down db-up db-migrate
	@$(GO_TEST_CMD) ./...

.PHONY: test-package
test-package: db-migrate
	@$(GO_TEST_CMD) $(P)

.PHONY: db-inject
db-inject:
	@zcat $(MYSQL_DUMP) | $(DOCKER_COMPOSE_CMD) exec -T database mysql -u$(MYSQL_USER) -p$(MYSQL_PASS) $(MYSQL_DB)

$(ENUMER):
	@go install github.com/dmarkham/enumer

$(GOLINT):
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.59.1

$(GOIMPORTS):
	@go install golang.org/x/tools/cmd/goimports

$(MOCKGEN):
	@go get -u github.com/golang/mock/mockgen

.PHONY: lint
lint: $(GOLINT) $(GOIMPORTS)
	@golangci-lint run -v

.PHONY: db-up
db-up: # Start the DB server
ifeq ($(IS_CI), 0)
	@$(DOCKER_COMPOSE_CMD) up -d database --remove-orphans
	@$(DOCKER_COMPOSE_CMD) run wait -c database:3306
endif

.PHONY: down
down:
	@$(DOCKER_COMPOSE_CMD) down

.PHONY: db-migrate
db-migrate: db-up
	@$(GO_RUN_CMD) scripts/migrate.go

.PHONY: db-cli
db-cli:
	@$(DOCKER_COMPOSE_CMD) exec database mysql -u$(MYSQL_USER) -p$(MYSQL_PASS) $(MYSQL_DB)

.PHONY: generate
generate: $(GOIMPORTS) $(ENUMER) $(MOCKGEN)
	@rm -rf ./mock/
	@$(GO_CMD) generate ./...
	@goimports -w ./mock

.PHOHY: goimports
goimports:
	@goimports -w ./
