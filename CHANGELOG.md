## [Unreleased]

### Changed

- **[breaking]** Remove the repeated Backend and abstracted it to it's own package
  ([Pull #60](https://github.com/cycloidio/terracost/pull/59))

## [0.4.4] _2021-09-31_

### Fixed

- Ignore providers related to constraints
  ([Pull #58](https://github.com/cycloidio/terracost/pull/58))
- Unexpected error when using supported/unsupported providers
  ([Pull #56](https://github.com/cycloidio/terracost/pull/56))

## [0.4.3] _2021-08-30_

### Changed

- Improved error returned when using unknown providers, empty terraform
  ([Pull #55](https://github.com/cycloidio/terracost/pull/55))

## [0.4.2] _2021-07-22_

### Fixed

- Add missing type/provider to resources in plans
  ([Pull #54](https://github.com/cycloidio/terracost/pull/54))

## [0.4.1] _2021-07-15_

### Fixed

- Correct calculation for planned cost in resource diff
  ([Issue #51](https://github.com/cycloidio/terracost/issues/51))

## [0.4.0] _2021-07-13_

### Added

- Include currency in resource component cost estimation
  ([Issue #48](https://github.com/cycloidio/terracost/issues/48))

## [0.3.0] _2021-06-01_

### Added

- Support estimation of a directory of HCL files
  ([Issue #29](https://github.com/cycloidio/terracost/issues/29))

## [0.2.0] _2021-05-14_

### Added

- Expose `provider` and `type` fields for estimated resources
  ([Issue #42](https://github.com/cycloidio/terracost/issues/42))

## [0.1.1] _2021-04-29_

### Fixed

- Add support for child modules in Terraform plan files
  ([Issue #37](https://github.com/cycloidio/terracost/issues/37))
