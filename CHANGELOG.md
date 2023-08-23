## [Unreleased]

### Added

- AWS support for `aws_nat_gateway`
  ([Pull #110](https://github.com/cycloidio/terracost/pull/110))
- AWS support for `aws_eks_cluster`, `aws_eks_node_group`, `aws_efs_file_system`
  ([Pull #97](https://github.com/cycloidio/terracost/pull/97))
- Support for remote module references
  ([Issue #88](https://github.com/cycloidio/terracost/issues/88))
- Added 'Usage' support for those options that are not from configuration but from usage of the resource
  ([Issue #96](https://github.com/cycloidio/terracost/issues/96))
- Added 'Usage' attribute to the cost.Component and query.Component
  ([Issue #10](https://github.com/cycloidio/terracost/issues/100))
- Plan estimation now supports references to other resources (like in ASG or EKS)
  ([Pull #105](https://github.com/cycloidio/terracost/pull/105))

### Fixed

- Plan variables where forced to be `string` and if not was failing
  ([Issue #86](https://github.com/cycloidio/terracost/issues/86))
- Added weak type conversion for values when calculation price from the user inputs
  ([Pull #89](https://github.com/cycloidio/terracost/pull/89))
- Resource.Index is now of type `interface{}` as it can be `int` or `string`
  ([Pull #90](https://github.com/cycloidio/terracost/pull/90))
- If the AWS provider is defined without region it'll use `us-east-1`
  ([Pull #94](https://github.com/cycloidio/terracost/pull/94))
- The documentation to install is has been updated to display the needed `replace` of `terraform`
  ([Issue #90](https://github.com/cycloidio/terracost/issues/90)), ([Issue #74](https://github.com/cycloidio/terracost/issues/74))

## [0.5.1] _2023-03-08_

### Added

- AWS support for `aws_autoscaling_group`, `aws_launch_template`, `aws_launch_configuration`
  ([Pull #80](https://github.com/cycloidio/terracost/pull/80))
- AWS support for `aws_eip`, `aws_aws_elasticache_cluster`, `aws_elasticache_replication_group`
  ([Pull #68](https://github.com/cycloidio/terracost/pull/68))

### Changed

- AWS ingester change to keep consistency between name from CSV
  ([Pull #68](https://github.com/cycloidio/terracost/pull/68))
- Internal signature for computing resources price, now also have the map of all the other resources to compute so it can access them when referenced
  ([Pull #68](https://github.com/cycloidio/terracost/pull/82))

## [0.5.0] _2022-01-18_

### Added

- Google simple implementation with support for `compute_instances`
  ([Issue #2](https://github.com/cycloidio/terracost/issues/2))
- AzureRM simple implementation with support for `virtual_machine`(just linux) and `linux_virtual_machine`
  ([Issue #3](https://github.com/cycloidio/terracost/issues/3))

### Changed

- **[breaking]** Remove the repeated Backend and abstracted it to it's own package
  ([Pull #60](https://github.com/cycloidio/terracost/pull/59))
- No cost returned when no prior exist in plan
  ([Pull #61](https://github.com/cycloidio/terracost/pull/61))

## [0.4.4] _2021-08-31_

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
