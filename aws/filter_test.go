package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
)

func TestMinimalFilter(t *testing.T) {
	t.Run("Allowed", func(t *testing.T) {
		pps := []*price.WithProduct{
			{Product: &product.Product{Service: "AmazonEC2", Family: "Storage"}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "System Operation"}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "Compute Instance", Attributes: map[string]string{
				"capacitystatus":  "Used",
				"operatingSystem": "Linux",
				"preInstalledSw":  "NA",
				"tenancy":         "Shared",
			}}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "Compute Instance", Attributes: map[string]string{
				"capacitystatus":  "Used",
				"operatingSystem": "Linux",
				"preInstalledSw":  "NA",
				"tenancy":         "Dedicated",
			}}},
			{Product: &product.Product{Service: "AmazonRDS", Family: "Database Instance"}},
			{Product: &product.Product{Service: "AmazonRDS", Family: "Database Storage"}},
			{Product: &product.Product{Service: "AmazonRDS", Family: "Provisioned IOPS"}},
		}

		for i, pp := range pps {
			assert.True(t, MinimalFilter(pp), "case %d", i)
		}
	})

	t.Run("Skipped", func(t *testing.T) {
		pps := []*price.WithProduct{
			{Product: &product.Product{Service: "AmazonEC2", Family: "Compute Instance (bare metal)"}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "NAT Gateway"}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "Fee"}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "Compute Instance", Attributes: map[string]string{
				"capacitystatus":  "Used",
				"operatingSystem": "Linux",
				"preInstalledSw":  "NA",
				"tenancy":         "Host",
			}}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "Compute Instance", Attributes: map[string]string{
				"capacitystatus":  "Used",
				"operatingSystem": "SUSE",
				"preInstalledSw":  "NA",
				"tenancy":         "Shared",
			}}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "Compute Instance", Attributes: map[string]string{
				"capacitystatus":  "Used",
				"operatingSystem": "Linux",
				"preInstalledSw":  "SQL Server",
				"tenancy":         "Shared",
			}}},
			{Product: &product.Product{Service: "AmazonEC2", Family: "Compute Instance", Attributes: map[string]string{
				"capacitystatus":  "AllocatedCapacityReservation",
				"operatingSystem": "Linux",
				"preInstalledSw":  "NA",
				"tenancy":         "Shared",
			}}},
			{Product: &product.Product{Service: "AmazonRDS", Family: "CPU Credits"}},
			{Product: &product.Product{Service: "AmazonRDS", Family: "Performance Insights"}},
			{Product: &product.Product{Service: "AmazonRDS", Family: "RDSProxy"}},
			{Product: &product.Product{Service: "AmazonRDS", Family: "Serverless"}},
		}

		for i, pp := range pps {
			assert.False(t, MinimalFilter(pp), "case %d", i)
		}
	})
}
