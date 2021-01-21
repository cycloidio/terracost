// Package terraform includes functionality related to reading Terraform plan files. The plan schema is defined
// here according to the JSON output format described at https://www.terraform.io/docs/internals/json-format.html
//
// The found resources are then transformed into query.Resource that can be utilized further.
package terraform
