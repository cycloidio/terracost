package costestimation

import (
	"context"
	"io"

	"github.com/cycloidio/cost-estimation/aws"
	"github.com/cycloidio/cost-estimation/cost"
	"github.com/cycloidio/cost-estimation/terraform"
)

// EstimateTerraformPlan is a helper function that reads a Terraform plan using the provided io.Reader,
// generates the prior and planned cost.State, and then creates a cost.Plan from them that is returned.
// It uses the Backend to retrieve the pricing data.
func EstimateTerraformPlan(ctx context.Context, backend Backend, r io.Reader, options ...terraform.Option) (*cost.Plan, error) {
	if len(options) == 0 {
		options = []terraform.Option{
			terraform.WithProvider("aws", aws.NewTerraformProvider),
		}
	}

	tfplan := terraform.NewPlan(options...)
	if err := tfplan.Read(r); err != nil {
		return nil, err
	}

	prior, err := cost.NewState(ctx, backend, tfplan.ExtractPriorQueries())
	if err != nil {
		return nil, err
	}
	planned, err := cost.NewState(ctx, backend, tfplan.ExtractPlannedQueries())
	if err != nil {
		return nil, err
	}

	return cost.NewPlan(prior, planned), nil
}
