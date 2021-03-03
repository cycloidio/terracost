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

## Adding a new resource

### AWS

1. Familiarize yourself with the official AWS pricing page for the service as well as the Terraform documentation for the resource you want to add. Note all factors that influence the cost.
2. Download and familiarize yourself with the pricing data CSV. This can be done by first checking the [index.json](https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/index.json), finding the respective service under the `offers` key and downloading the file at the URL under the `currentVersionUrl` (replace `json` with `csv`).
3. Find the names of all columns that contain relevant cost factors and check that the `aws/field/field.go` file contains them - add them if this is not the case. The constant name should be a correct Go identifier, while the comment should contain the name as it appears in the CSV file.
4. Run `go generate ./...` to regenerate the field list.
5. Create a new file with the name of the Terraform resource (without the `aws` prefix), e.g. for `aws_db_instance` it would be `db_instance.go`. It should include two new structs: `Resource` (that is an intermediate struct containing only the relevant cost factors) and `resourceValues` (that directly represents the values from the Terraform resource.) Additionally, the `Resource` struct must implement the `Components` method that returns `[]query.Component`. See the other existing resources for inspiration.
6. Write tests for your resource. As before, check the other existing test files for inspiration.
7. Test and make sure that estimating your resource works.
8. Open a PR with the changes and please try to provide as much information as possible, especially: description of all the cost factors that the PR uses, links to Terraform docs and AWS pricing page, examples of a Terraform file and the resulting estimation.

## Adding a new provider/backend

**Please be aware that, at the moment, Cycloid only supports MySQL as a backend and AWS as cloud provider.** Based on this, please refrain from making contributions that add a new backend or cloud provider as we cannot guarantee they'd be merged and/or supported. To make improvements in this area, please instead open an appropriate issue so that we can discuss it and provide any necessary guidance.
