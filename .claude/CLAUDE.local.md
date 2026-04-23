# Personal Claude Instructions for Lukasz

## Private Rules

- Prefer the `Makefile` targets in this repo (`make test`, `make lint`, `make generate`). Tests need MySQL running via docker-compose, so a bare `go test ./...` will fail for most packages.
- When I ask to "add a resource" for a provider, the work is almost always in two places: `<provider>/terraform/` (resource → product/query mapping) and the provider's top-level filter/options/services file. Skim an existing similar resource before starting.
- Never reach for `float64` in pricing math. Use `shopspring/decimal`. I'd rather a verbose `decimal.NewFromFloat(...).Mul(...)` chain than a "temporary" float.
- When mocks look stale after editing an interface, run `make generate` and commit the regenerated `mock/` tree in the same commit — don't leave them out of sync.
