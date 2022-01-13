# AWS

To know how to price each resource it's good to check the [AWS Price Calculator](https://calculator.aws/#/estimate) and the CSV that we
use for AWS has [this](https://docs.aws.amazon.com/cur/latest/userguide/product-columns.html) columns and format.

## Adding new resources

1. Familiarize yourself with the official AWS pricing page for the service as well as the Terraform documentation for the resource you want to add. Note all factors that influence the cost.
2. Download and familiarize yourself with the pricing data CSV. This can be done by first checking the [index.json](https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/index.json), finding the respective service under the `offers` key and downloading the file at the URL under the `currentVersionUrl` (replace `json` with `csv`).
3. Find the names of all columns that contain relevant cost factors and check that the `aws/field/field.go` file contains them - add them if this is not the case and also to the `aws/ingester.go` so it's categorized to the right entity (Price or Product). The constant name should be a correct Go identifier, while the comment should contain the variable name as it appears in `aws/field/field.go`.
4. Run `make generate` to regenerate the field list.
5. Create a new file in the `aws/terraform` directory with the name of the Terraform resource (without the `aws` prefix), e.g. for `aws_db_instance` it would be `db_instance.go`. It should include two new structs: `Resource` (that is an intermediate struct containing only the relevant cost factors) and `resourceValues` (that directly represents the values from the Terraform resource.) Additionally, the `Resource` struct must implement the `Components` method that returns `[]query.Component`. See the other existing resources for inspiration.
6. Add the terraform resource to the `aws/terraform/provider.go`  on the `ResourceComponents`
7. Write tests for your resource. As before, check the other existing test files for inspiration.
8. Test and make sure that estimating your resource works.
9. Open a PR with the changes and please try to provide as much information as possible, especially: description of all the cost factors that the PR uses, links to Terraform docs and AWS pricing page, examples of a Terraform file and the resulting estimation.

## List of supported resources and attributes

* [`aws_instance`](#aws_instance)
* [`aws_db_instance`](#aws_db_instance)
* [`aws_ebs_volume`](#aws_ebs_volume)
* [`aws_lb/aws_alb`](#aws_lb--aws_alb)
* [`aws_elb`](#aws_elb)

### `aws_instance`

#### Cost factors

* Location
* Instance type
* Tenancy - only "shared" and "dedicated"
* Operating system - currently only Linux supported, every instance is treated as a Linux instance
* Pre-installed S/W - currently not supported, the value of "NA" is used instead
* Storage - see more in the `aws_ebs_volume` entry

#### Additional notes

* Only "On Demand" instances are supported.
* Only compute and storage costs are estimated. GPU, monitoring, etc. are not taken into account.
* Uptime of 730 hours in a month (non-stop) is assumed.

### `aws_db_instance`

#### Cost factors

* Location
* Instance class
* Database engine and edition
* License model - "License included" or "Bring your own license"
* Deployment option - "Single-AZ" or "Multi-AZ"
* Allocated storage
* Storage type - "Magnetic" (standard), "Provisioned IOPS" (io1), "General Purpose" (gp2)
* Provisioned IOPS - only for this type of storage; 100 by default

#### Additional notes

* Only "On Demand" database instances are supported.
* Uptime of 730 hours in a month (non-stop) is assumed.

### `aws_ebs_volume`

#### Cost factors

* Location
* Volume type - "gp2" by default
* Volume size - 8GB by default
* Provisioned IOPS - only for "io1" and "io2" volume types; 100 by default


### `aws_lb` / `aws_alb`

#### Cost factors

* Location
* Load balancer type - "application" by default

#### Additional notes

* Cost of Load Balancer Capacity Units (LCU's) per hour is not estimated.

### `aws_elb`

#### Cost factors

* Location

#### Additional notes

* Data transfer usage cost is not estimated.
