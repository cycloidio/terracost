# Claude Code Instructions for Terracost

## Project Overview

Terracost is a Go library maintained by Cycloid for estimating the cost of Terraform-managed infrastructure. It parses a Terraform plan or HCL config, looks up prices for each resource in a MySQL-compatible database that has been pre-populated from cloud vendor pricing APIs, and returns per-resource and total cost estimates.

It is a **library, not a binary** — downstream programs (APIs, CLIs) import it and supply their own `*sql.DB`, credentials, plan JSON, etc. The two main entry points are:

- `terracost.EstimateTerraformPlan` / `EstimateHCL` in `estimation.go` — produce a cost estimate from user input.
- `terracost.IngestPricing` in `ingestion.go` — populate the pricing DB from a vendor ingester (AWS/Google/AzureRM).

Module path: `github.com/cycloidio/terracost`.

## Core Rules

- **Match existing conventions.** This is a long-lived Cycloid library with a consistent house style; follow what surrounding files already do (logging, error wrapping, test layout) rather than importing conventions from other projects.
- **Prices are money — use `decimal`.** All cost and price math goes through `github.com/shopspring/decimal`. Never convert to `float64` for arithmetic, even "just for a log line".
- **Provider code lives under its provider package.** AWS-specific resources, filters, and regions go under `aws/`; same for `google/` and `azurerm/`. Don't push provider details up into `cost/`, `price/`, `product/`, `query/`, or `terraform/`.
- **The `terraform/` package parses plans/HCL — it is not the HashiCorp fork.** The HashiCorp fork is a separate repo pulled in via `replace github.com/hashicorp/terraform => github.com/cycloidio/terraform`. Don't confuse the two when the user says "terraform".
- **English only**, no emojis, no trailing periods on comments, state the *why* not the *what*.

## Repository Layout

```
terracost/
├── estimation.go, ingestion.go, provider.go, doc.go  # public top-level API
├── aws/           # AWS ingester, filters, region codes, TF resource → product mapping
│   └── terraform/ # AWS-specific terraform resource handlers
├── google/        # Google Cloud ingester + terraform handlers
├── azurerm/       # Azure ingester + terraform handlers
├── terraform/     # Plan/HCL parsing, schema, provider detection (NOT the HashiCorp fork)
├── cost/          # Cost/Component/Plan/Resource/State domain types
├── price/         # Price entity + repository interface
├── product/       # Product entity (a priced SKU) + repository interface
├── query/         # Query abstraction fed into provider implementations
├── usage/         # User-supplied usage overrides (e.g. expected monthly GB)
├── backend/       # Backend interface that bundles Price+Product repositories
├── mysql/         # MySQL implementation of backend + migrations + testdata dump
├── log/           # Thin logrus wrapper
├── util/          # Small shared helpers
├── mock/          # gomock-generated mocks (regenerated via `make generate`)
├── testutil/      # Shared test helpers
├── e2e/           # End-to-end tests (hit the real DB via docker-compose)
├── examples/      # Standalone example module showing library usage
├── scripts/       # migrate.go and other ops scripts
├── tools/         # Developer tooling
├── docs/          # Per-provider supported-resource docs
├── docker-compose.yml
└── Makefile
```

Single long-lived branch: `master`. Feature work happens on short-lived branches (see `git branch -r`) merged via PR.

## Build, Test & Code Quality

This repo **does** use a Makefile — unlike some Cycloid repos, prefer the Make targets over invoking `go` directly, because tests need a running MySQL via docker-compose.

```bash
make test            # docker-compose up MySQL, run migrations, then `go test ./...`
make test-package P=./aws/...   # run one subtree (assumes DB is already migrated)
make lint            # golangci-lint (installs to ./bin on first run)
make generate        # regenerate mocks and enumer stringers (rm -rf ./mock first)
make goimports       # goimports -w ./
make db-up           # start just the MySQL container
make db-migrate      # run scripts/migrate.go against the running DB
make db-inject       # load the pricing dump from mysql/testdata/*.sql.gz
make db-cli          # open a MySQL shell inside the container
make down            # docker-compose down
make ci              # what CI runs: lint + go test ./...
```

Plain `go test ./...` works only for packages that don't touch MySQL — most package tests assume the DB is up and migrated. When in doubt, `make test`.

Generated files (`mock/**`, `*_enumer.go`, anything a `go:generate` directive produces) must never be hand-edited — change the source and re-run `make generate`.

## Code Conventions

- **Errors**: wrap with `fmt.Errorf("...: %w", err)` and use `errors.Is/As`. There are a few package-level sentinel errors (e.g. `aws.ErrNoPrice`) — reuse them rather than inventing new ones.
- **Logging**: `github.com/sirupsen/logrus`, usually accessed via the thin `log/` package. Do not introduce `slog` or `hclog`.
- **Money math**: `github.com/shopspring/decimal`, never `float64`.
- **Context**: public functions that do I/O (DB, HTTP, ingesters) take `context.Context` as the first arg; pure computations don't.
- **Comments**: godoc-style, starting with the identifier name (`// Ingester fetches ...`).
- **Tests**: table-driven with `t.Run(name, func(t *testing.T) { ... })`. Mocks via `github.com/golang/go-mock/gomock` generated into `mock/`. DB tests use `github.com/DATA-DOG/go-sqlmock` for unit tests and real MySQL (via docker-compose) for integration/e2e. Fixtures live under `testdata/` directories next to the code under test.
- **Mocks**: add a `//go:generate mockgen ...` directive to the interface's file; `make generate` wipes and regenerates the whole `mock/` tree.

## Working with the Terraform Fork Dependency

Terracost depends on a Cycloid fork of HashiCorp Terraform, pinned in `go.mod`:

```
replace github.com/hashicorp/terraform => github.com/cycloidio/terraform v1.4.6-cy
```

Downstream consumers must apply the same replace (the README shows the exact `go mod edit` command). If something imported from `github.com/hashicorp/terraform/...` is missing or internal, the fix usually lives in the `cycloidio/terraform` repo (another checkout under `~/go/src/github.com/cycloidio/terraform`), not here. Don't vendor upstream Terraform types into this repo as a workaround.

## Things NOT to Do

- Don't change the module path.
- Don't introduce `float64` into pricing or cost math.
- Don't hand-edit files under `mock/` or any `_enumer.go` / generated file — re-run `make generate`.
- Don't add a provider-specific dependency to a non-provider package. Keep `aws/`, `google/`, `azurerm/` self-contained.
- Don't bypass the Makefile for tests that need MySQL — the `docker-compose` step and `scripts/migrate.go` are doing real work.
- Don't add new dependencies casually; this is a library that downstreams import, so every new dep affects their dependency tree.
