package terracost

import (
	"context"
	"io"

	"github.com/spf13/afero"

	"github.com/cycloidio/terracost/aws"
	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/terraform"
)

// EstimateTerraformPlan is a helper function that reads a Terraform plan using the provided io.Reader,
// generates the prior and planned cost.State, and then creates a cost.Plan from them that is returned.
// It uses the Backend to retrieve the pricing data.
func EstimateTerraformPlan(ctx context.Context, backend Backend, plan io.Reader, providerInitializers ...terraform.ProviderInitializer) (*cost.Plan, error) {
	if len(providerInitializers) == 0 {
		providerInitializers = []terraform.ProviderInitializer{
			aws.TerraformProviderInitializer,
		}
	}

	tfplan := terraform.NewPlan(providerInitializers...)
	if err := tfplan.Read(plan); err != nil {
		return nil, err
	}

	priorQueries, err := tfplan.ExtractPriorQueries()
	if err != nil && err != terraform.ErrNoQueries {
		return nil, err
	}
	prior, err := cost.NewState(ctx, backend, priorQueries)
	if err != nil {
		return nil, err
	}

	plannedQueries, err := tfplan.ExtractPlannedQueries()
	if err != nil {
		return nil, err
	}
	planned, err := cost.NewState(ctx, backend, plannedQueries)
	if err != nil {
		return nil, err
	}

	return cost.NewPlan(prior, planned), nil
}

// EstimateHCL is a helper function that recursively reads Terraform modules from a directory at the
// given path and generates a planned cost.State that is returned wrapped in a cost.Plan.
// It uses the Backend to retrieve the pricing data.
func EstimateHCL(ctx context.Context, backend Backend, fs afero.Fs, path string, providerInitializers ...terraform.ProviderInitializer) (*cost.Plan, error) {
	if len(providerInitializers) == 0 {
		providerInitializers = []terraform.ProviderInitializer{
			aws.TerraformProviderInitializer,
		}
	}

	plannedQueries, err := terraform.ExtractQueriesFromHCL(fs, providerInitializers, path)
	if err != nil {
		return nil, err
	}
	planned, err := cost.NewState(ctx, backend, plannedQueries)
	if err != nil {
		return nil, err
	}

	return cost.NewPlan(nil, planned), nil
}
