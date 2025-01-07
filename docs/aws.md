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

<!--
for i in $(grep 'case ' aws/terraform/provider.go | sed -E 's/.*case[^"]+//;s/[",:]//g');do
  shortname=$(echo $i| sed -E 's/^aws_//')
  echo '* [`'$i'`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/'$shortname')';
done
-->

* [`aws_instance`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance)
* [`aws_autoscaling_group`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/autoscaling_group)
* [`aws_db_instance`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_instance)
* [`aws_ebs_volume`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ebs_volume)
* [`aws_efs_file_system`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/efs_file_system)
* [`aws_elasticache_cluster`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/elasticache_cluster)
* [`aws_elasticache_replication_group`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/elasticache_replication_group)
* [`aws_eip`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eip)
* [`aws_elb`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/elb)
* [`aws_eks_cluster`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster)
* [`aws_eks_node_group`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_node_group)
* [`aws_fsx_lustre_file_system`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/fsx_lustre_file_system)
* [`aws_fsx_ontap_file_system`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/fsx_ontap_file_system)
* [`aws_fsx_openzfs_file_system`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/fsx_openzfs_file_system)
* [`aws_fsx_windows_file_system`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/fsx_windows_file_system)
* [`aws_lb`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lb)
* [`aws_alb`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/alb)
* [`aws_nat_gateway`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/nat_gateway)