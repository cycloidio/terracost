IS_CI ?= 0

DOCKER_COMPOSE_CMD := docker-compose
GO_CMD := go
GO_TEST_CMD := $(GO_CMD) test
GO_RUN_CMD := $(GO_CMD) run

MYSQL_USER := root
MYSQL_PASS := terracost
MYSQL_DB := terracost_test
MYSQL_DUMP ?= mysql/testdata/2021-08-12-pricing.sql.gz

# The ci rule is voluntarily simple because
# the CI requires to run the command separately
# and not all at once, otherwise it wouldn't work
.PHONY: ci
ci: lint
	@$(GO_TEST_CMD) ./...

.PHONY: test
test: lint down db-up db-migrate db-inject
	@$(GO_TEST_CMD) ./...

.PHONY: db-inject
db-inject:
	@zcat $(MYSQL_DUMP) | $(DOCKER_COMPOSE_CMD) exec -T database mysql -u$(MYSQL_USER) -p$(MYSQL_PASS) $(MYSQL_DB)

$(ENUMER):
	@go install github.com/dmarkham/enumer

$(GOLINT):
	@go install golang.org/x/lint/golint

$(GOIMPORTS):
	@go install golang.org/x/tools/cmd/goimports

.PHONY: lint
lint: $(GOLINT)
	@golint -set_exit_status ./... && test -z "`$(GO_CMD) list -f {{.Dir}} ./... | xargs goimports -l | tee /dev/stderr`"

.PHONY: db-up
db-up: # Start the DB server
ifeq ($(IS_CI), 0)
	@$(DOCKER_COMPOSE_CMD) up -d database
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
generate: $(GOIMPORTS) $(ENUMER)
	@rm -rf ./mock/
	@$(GO_CMD) generate ./...
	@goimports -w ./mock

.PHOHY: goimports
goimports:
	@goimports -w ./
