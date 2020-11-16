package terraform

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/cost-estimation/price"
	"github.com/cycloidio/cost-estimation/product"
	"github.com/cycloidio/cost-estimation/query"
	"github.com/cycloidio/cost-estimation/terraform"
	"github.com/cycloidio/cost-estimation/util"
)

func TestDBInstance_Components(t *testing.T) {
	p, err := NewProvider("aws", "eu-west-3")
	require.NoError(t, err)

	t.Run("DefaultValues", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_db_instance.test",
			Type:         "aws_db_instance",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"instance_class":    "db.t2.xlarge",
				"allocated_storage": float64(42),
				"engine":            "postgres",
			},
		}

		expected := []query.Component{
			{
				Name:           "Database instance",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Single-AZ", "db.t2.xlarge"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Instance"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "instanceType", Value: util.StringPtr("db.t2.xlarge")},
						{Key: "deploymentOption", Value: util.StringPtr("Single-AZ")},
						{Key: "databaseEngine", Value: util.StringPtr("PostgreSQL")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "purchaseOption", Value: util.StringPtr("on_demand")},
					},
				},
			},
			{
				Name:            "Database storage",
				MonthlyQuantity: decimal.NewFromFloat(42),
				Unit:            "GB",
				Details:         []string{"General Purpose"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Storage"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "deploymentOption", Value: util.StringPtr("Single-AZ")},
						{Key: "volumeType", Value: util.StringPtr("General Purpose")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("IoStorageType", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_db_instance.test",
			Type:         "aws_db_instance",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"instance_class":    "db.t2.xlarge",
				"storage_type":      "io1",
				"iops":              float64(200),
				"allocated_storage": float64(42),
				"engine":            "postgres",
			},
		}

		expected := []query.Component{
			{
				Name:           "Database instance",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Single-AZ", "db.t2.xlarge"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Instance"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "instanceType", Value: util.StringPtr("db.t2.xlarge")},
						{Key: "deploymentOption", Value: util.StringPtr("Single-AZ")},
						{Key: "databaseEngine", Value: util.StringPtr("PostgreSQL")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "purchaseOption", Value: util.StringPtr("on_demand")},
					},
				},
			},
			{
				Name:            "Database storage",
				MonthlyQuantity: decimal.NewFromFloat(42),
				Unit:            "GB",
				Details:         []string{"Provisioned IOPS"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Storage"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "deploymentOption", Value: util.StringPtr("Single-AZ")},
						{Key: "volumeType", Value: util.StringPtr("Provisioned IOPS")},
					},
				},
			},
			{
				Name:            "Database IOPS",
				MonthlyQuantity: decimal.NewFromFloat(200),
				Unit:            "IOPS",
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Provisioned IOPS"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "deploymentOption", Value: util.StringPtr("Single-AZ")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("WithLicenseModelMultiAZ", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_db_instance.test",
			Type:         "aws_db_instance",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"instance_class":    "db.t2.xlarge",
				"allocated_storage": float64(42),
				"engine":            "oracle-se1",
				"license_model":     "bring-your-own-license",
				"multi_az":          true,
			},
		}

		expected := []query.Component{
			{
				Name:           "Database instance",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Multi-AZ", "db.t2.xlarge"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Instance"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "instanceType", Value: util.StringPtr("db.t2.xlarge")},
						{Key: "deploymentOption", Value: util.StringPtr("Multi-AZ")},
						{Key: "databaseEngine", Value: util.StringPtr("Oracle")},
						{Key: "databaseEdition", Value: util.StringPtr("Standard One")},
						{Key: "licenseModel", Value: util.StringPtr("Bring your own license")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "purchaseOption", Value: util.StringPtr("on_demand")},
					},
				},
			},
			{
				Name:            "Database storage",
				MonthlyQuantity: decimal.NewFromFloat(42),
				Unit:            "GB",
				Details:         []string{"General Purpose"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Storage"),
					Location: util.StringPtr("eu-west-3"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "deploymentOption", Value: util.StringPtr("Multi-AZ")},
						{Key: "volumeType", Value: util.StringPtr("General Purpose")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})
}
