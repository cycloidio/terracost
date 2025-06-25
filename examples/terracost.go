package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cycloidio/terracost"
	"github.com/cycloidio/terracost/aws"
	"github.com/cycloidio/terracost/aws/region"
	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/azurerm"
	azuretf "github.com/cycloidio/terracost/azurerm/terraform"
	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/google"
	googletf "github.com/cycloidio/terracost/google/terraform"
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
	flagIngest               bool   = false
	flagIngestMinimal        bool   = true
	flagIngestRegion         string = ""
	defaultAWSRegion         string = "eu-west-1"
	defaultAzureRegion       string = "francecentral"
	defaultGCPRegion         string = "europe-west1-b"
	flagestimatePlan         string = ""
	flagestimateHCL          string = ""
	flagProvider             string = "aws"
	googleCredentialFilePath string = "/tmp/credentials.json"
)

func main() {

	flag.Usage = helpUsage
	flag.BoolVar(&flagIngest, "ingest", flagIngest, "Run price ingester")
	flag.BoolVar(&flagIngestMinimal, "ingest-minimal", flagIngestMinimal, "Minimal ingest")
	flag.StringVar(&flagIngestRegion, "ingest-region", flagIngestRegion, "Region used to ingest")
	flag.StringVar(&flagestimatePlan, "estimate-plan", flagestimatePlan, "terraform-plan.json file path to estimate (example: ./terraform-plan.json)")
	flag.StringVar(&flagestimateHCL, "estimate-hcl", flagestimateHCL, "terraform HCL code path to estimate (example: ../testdata/aws/stack-aws)")
	flag.StringVar(&flagProvider, "provider", flagProvider, "Terraform provider used [aws|azure|gcp]")
	flag.StringVar(&googleCredentialFilePath, "google-cred-file", googleCredentialFilePath, "GCP JSON credential file path (/tmp/credentials.json)")

	flag.Parse()

	// get command line args
	args := flag.Args()
	if len(args) == 0 {
	}

	if !flagIngest && flagestimatePlan == "" && flagestimateHCL == "" {
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

	if flagIngest {
		ingestRegion := flagIngestRegion
		if ingestRegion == "" {
			if flagProvider == "aws" {
				ingestRegion = defaultAWSRegion
			} else if flagProvider == "azure" {
				ingestRegion = defaultAzureRegion
			} else if flagProvider == "gcp" {
				ingestRegion = defaultGCPRegion
			}
		}
		ingest(flagProvider, ingestRegion, backend, db)
	}

	if flagestimatePlan != "" {
		estimatePlan(flagestimatePlan, backend)
	}

	if flagestimateHCL != "" {
		estimateHCL(flagestimateHCL, flagProvider, backend)
	}

}
func ingest(flagProvider string, region string, backend *mysql.Backend, db *sql.DB) {
	fmt.Printf("Ingestion %s\n", flagProvider)

	err := mysql.Migrate(context.Background(), db, "pricing_migrations")
	if err != nil && !strings.Contains(err.Error(), "Error 1050") {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	if flagProvider == "aws" {
		// Ingest supported services pricing data into the database
		for _, s := range aws.GetSupportedServices() {
			fmt.Printf("[%s] Ingestion\n", s)
			op := []aws.Option{}
			if flagIngestMinimal {
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
	} else if flagProvider == "azure" {
		for _, s := range azurerm.GetSupportedServices() {
			fmt.Printf("[%s] Ingestion\n", s)
			op := []azurerm.Option{}
			if flagIngestMinimal {
				op = append(op, azurerm.WithIngestionFilter(azurerm.MinimalFilter))
			}
			ingester, err := azurerm.NewIngester(context.Background(), s, region, op...)
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
	} else if flagProvider == "gcp" {
		for _, s := range google.GetSupportedServices() {
			fmt.Printf("[%s] Ingestion\n", s)
			op := []google.Option{}
			if flagIngestMinimal {
				op = append(op, google.WithIngestionFilter(google.MinimalFilter))
			}

			// Read a GCP access credentials json file
			// https://developers.google.com/workspace/guides/create-credentials?hl=en
			file, err := ioutil.ReadFile(googleCredentialFilePath)
			if err != nil {
				fmt.Printf("Failed to open google credential file: %s\n", err)
			}

			type GoogleCredential struct {
				ProjectID string `json:"project_id"`
			}

			// Define a variable to hold the data
			var credential GoogleCredential
			err = json.Unmarshal(file, &credential)
			if err != nil {
				fmt.Printf("Failed to decode JSON: %s\n", err)
			}

			googleProject := credential.ProjectID
			googleCredentialJSON := file

			ingester, err := google.NewIngester(context.Background(), googleCredentialJSON, s, googleProject, region, op...)

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
	var terraformProviderInitializer = terraform.ProviderInitializer{}
	if provider == "aws" {
		terraformProviderInitializer = terraform.ProviderInitializer{
			MatchNames: []string{"aws", fmt.Sprintf("registry.terraform.io/hashicorp/%s", "aws")},
			Provider: func(config map[string]interface{}) (terraform.Provider, error) {
				r, ok := config["region"]
				if !ok {
					return nil, nil
				}
				regCode := region.Code(r.(string))
				return awstf.NewProvider(provider, regCode)
			},
		}

	} else if provider == "azure" {
		terraformProviderInitializer = terraform.ProviderInitializer{
			MatchNames: []string{"azurerm", fmt.Sprintf("registry.terraform.io/hashicorp/%s", "azurerm")},
			Provider: func(config map[string]interface{}) (terraform.Provider, error) {
				return azuretf.NewProvider("azurerm")
			},
		}
	} else if provider == "gcp" {
		terraformProviderInitializer = terraform.ProviderInitializer{
			MatchNames: []string{"google", fmt.Sprintf("registry.terraform.io/hashicorp/%s", "google")},
			Provider: func(config map[string]interface{}) (terraform.Provider, error) {
				r, ok := config["region"]
				if !ok {
					return nil, nil
				}
				return googletf.NewProvider(provider, r.(string))
			},
		}
	}
	// terraform HCL directory
	debugEnabled := true
	planhcl, err := terracost.EstimateHCL(context.Background(), backend, nil, path, "", false, 0, usage.Default, debugEnabled, terraformProviderInitializer)

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
