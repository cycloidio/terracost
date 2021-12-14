# Contributing Guidelines

Cycloid Team is glad to see you contributing to this project ! In this document, we will provide you some guidelines in order to help get your contribution accepted.

## Reporting an issue

### Issues

When you find a bug in Terracost, it should be reported using [Github issues](https://github.com/cycloidio/terracost/issues). Please provide key information like your Operating System (OS), Go version and finally the version of the library that you're using.

## Submit a contribution

### Setup your git repository

If you want to contribute to an existing issue, you can start by _forking_ this repository, then clone your fork on your machine.

```shell
$ git clone https://github.com/<your-username>/terracost.git
$ cd terracost
```

In order to stay updated with the upstream, it's highly recommended to add `cycloidio/terracost` as a remote upstream.

```shell
$ git remote add upstream https://github.com/cycloidio/terracost.git
```

Do not forget to frequently update your fork with the upstream.

```shell
$ git fetch upstream --prune
$ git rebase upstream/master
```

### Test your submission

First-time setup of the development environment requires running the database migrations:

```shell
$ make db-migrate
```

Run all the tests:

```shell
$ make test
```

## General information

### What is a SKU?

SKU means Stock Keeping Unit (SKU) and it is and identifier such as a number or a (bar-)code for a single item in a retail company’s range. The SKU enables a product to be clearly identified and is important for a retailer’s inventory management. 

So each SKU it's a different product which has a related price.

### What does it mean to calculate the cost of a resource

We have to make a link between the Cloud Provider SKU and the configuration attributes of the resource, either HCL or the Plan.
This link is done manually by knowing what the Cloud Provider is pricing for and how it's declared on Terraform, on the `aws_instance` case for example:

* `instance_type`(`"t2.micro"`): Which at the end is a number of CPU+RAM of the instance
* `tenancy`(`"dedicated"`): Tenancy of the instance

We track other attributes and combination of them, for example the `root_block_device` configuration block is
mapped to a `aws_ebs_volume` so at the end we have a total price which is a composition of other blocks.

We also take sometimes assumptions that later on may be parametrized, in this case we assume that it's always using a Linux OS

### Importing

The importing process is done by reading the information for the Cloud Provider API (or file) and then reduce it down to a `product.Product` and `price.Price` which
are the main entities we store on the DB and query to get the Prices. Both of them have an `Attributes` which is a `map[string]interface{}` which is used
to store specific information that is unique for the Cloud Provider than then we can use to make dedicated queries when fetching the prices.

To minimize the import volume we always import filter by a Service and a Region/Location/Zone (each Cloud Provider has it's own names) and we also have a `{provider}.MinimalFilter` which would only import from that previous filter only the elements we know and use at the moment, this is useful to not flood the DB with unnecessary prices. This options are used when initializing the specific Importer.

## Adding a new resource

### AWS

To add more AWS resources first read the [documentation](docs/aws.md) we have about it.

### Google

To add more Google resources first read the [documentation](docs/google.md) we have about it.

### AzureRM

To add more AzureRM resources first read the [documentation](docs/azurerm.md) we have about it.

## Testing resources

All the providers use API to get the needed data from the prices so we can ingest it, the way we use for testing is to mock that API call by being able to initialize
the provider with a different endpoint (on Azure we need to use `"https://prices.azure.com/"`) and on the test we just create a new [`httptest.NewServer`](https://pkg.go.dev/net/http/httptest#NewServer) and mock the needed endpoints. This implementations are in `testutil/` and the data that they return on `testdata/`. With that we can mock
any API call we need and expect the same results.

## Adding a new provider/backend

**Please be aware that, at the moment, Cycloid only supports MySQL as a backend.** Based on this, please refrain from making contributions that add a new backend or cloud provider as we cannot guarantee they'd be merged and/or supported. To make improvements in this area, please instead open an appropriate issue so that we can discuss it and provide any necessary guidance.
