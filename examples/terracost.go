package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/aws"
	"github.com/cycloidio/terracost/aws/region"
	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/usage"
	_ "github.com/go-sql-driver/mysql"
)

func helpUsage() {
	fmt.Fprint(os.Stderr, "Terracost\n\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	flagAWSIngest           bool   = false
	flagAWSIngestRegion     string = "eu-west-1"
	flagAWSIngestMinimal    bool   = false
	flagestimatePlan        string = ""
	flagestimateHCL         string = ""
	flagestimateHCLProvider string = "aws"
)

func main() {

	flag.Usage = helpUsage
	flag.BoolVar(&flagAWSIngest, "ingest-aws", flagAWSIngest, "Run AWS price ingester")
	flag.BoolVar(&flagAWSIngestMinimal, "minimal", flagAWSIngestMinimal, "Import the minimum amount of prices")
	flag.StringVar(&flagAWSIngestRegion, "ingest-aws-region", flagAWSIngestRegion, "AWS region used to ingest")
	flag.StringVar(&flagestimatePlan, "estimate-plan", flagestimatePlan, "terraform-plan.json file path to estimate (example: ./terraform-plan.json)")
	flag.StringVar(&flagestimateHCL, "estimate-hcl", flagestimateHCL, "terraform HCL code path to estimate (example: ../testdata/aws/stack-aws)")
	flag.StringVar(&flagestimateHCLProvider, "estimate-hcl-provider", flagestimateHCLProvider, "Terraform provider used")

	flag.Parse()

	// get command line args
	args := flag.Args()
	if len(args) == 0 {
	}

	if !flagAWSIngest && flagestimatePlan == "" && flagestimateHCL == "" {
		helpUsage()
		os.Exit(0)
	}

	// Use your mysql access with MultiStatements
	db, err := sql.Open("mysql", "root:terracost@tcp(127.0.0.1:3306)/terracost_test?multiStatements=true")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	backend := mysql.NewBackend(db)

	if flagAWSIngest {
		awsIngest(flagAWSIngestRegion, backend, db)
	}

	if flagestimatePlan != "" {
		estimatePlan(flagestimatePlan, backend)
	}

	if flagestimateHCL != "" {
		estimateHCL(flagestimateHCL, flagestimateHCLProvider, backend)
	}

}
func awsIngest(region string, backend *mysql.Backend, db *sql.DB) {
	fmt.Printf("AWS Ingestion\n")

	err := mysql.Migrate(context.Background(), db, "pricing_migrations")
	if err != nil && !strings.Contains(err.Error(), "Error 1050") {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	// Ingest supported services pricing data into the database
	for _, s := range aws.GetSupportedServices() {
		fmt.Printf("[%s] Ingestion\n", s)
		op := []aws.Option{}
		if flagAWSIngestMinimal {
			op = append(op, aws.WithIngestionFilter(aws.MinimalFilter))
		}
		ingester, err := aws.NewIngester(s, region, op...)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}

		err = terracost.IngestPricing(context.Background(), backend, ingester)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}
}

func estimatePlan(path string, backend *mysql.Backend) {
	fmt.Printf("EstimateTerraformPlan\n")
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	plan, err := terracost.EstimateTerraformPlan(context.Background(), backend, file, usage.Default)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	estimateDisplay(plan)
}

func estimateHCL(path string, provider string, backend *mysql.Backend) {
	fmt.Printf("EstimateHCL %s\n", provider)

	// From HCL
	// Provider to use
	var terraformAWSProviderInitializer = terraform.ProviderInitializer{
		MatchNames: []string{provider, fmt.Sprintf("registry.terraform.io/hashicorp/%s", provider)},
		Provider: func(config map[string]interface{}) (terraform.Provider, error) {
			r, ok := config["region"]
			if !ok {
				return nil, nil
			}
			regCode := region.Code(r.(string))
			return awstf.NewProvider(provider, regCode)
		},
	}

	// terraform HCL directory

	planhcl, err := terracost.EstimateHCL(context.Background(), backend, nil, path, "", false, usage.Default, terraformAWSProviderInitializer)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	for _, p := range planhcl {
		estimateDisplay(p)
	}
}

func estimateDisplay(resourceDiff *cost.Plan) {
	for _, res := range resourceDiff.ResourceDifferences() {
		priorCost, err := res.PriorCost()
		if err != nil {
			fmt.Printf("PriorCost %s: %s\n", res.Address, err)
			continue
		}

		plannedCost, err := res.PlannedCost()
		if err != nil {
			fmt.Printf("PlannedCost %s: %s\n", res.Address, err)
			continue
		}
		fmt.Printf("%s: %s -> %s\n", res.Address, priorCost, plannedCost)
	}
}
