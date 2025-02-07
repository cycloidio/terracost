package terraform_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/testutil"
	"github.com/cycloidio/terracost/usage"
	"github.com/cycloidio/terracost/util"
)

func TestRDSCluster_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("RDSClusterMysql", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_rds_cluster.test",
			Type:         "aws_rds_cluster",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"engine":                  "aurora-mysql",
				"backup_retention_period": 15,
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Storage",
				MonthlyQuantity: decimal.NewFromFloat(50),
				Unit:            "GB-Mo",
				Details:         []string{"Storage"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", ValueRegex: util.StringPtr("Any")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:StorageUsage$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB-Mo"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "I/O requests",
				MonthlyQuantity: decimal.NewFromFloat(21024000),
				Unit:            "IOs",
				Details:         []string{"I/O requests"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("System Operation"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", ValueRegex: util.StringPtr("Any")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:StorageIOUsage$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("IOs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Backup storage",
				MonthlyQuantity: decimal.NewFromFloat(840),
				Unit:            "GB-Mo",
				Details:         []string{"Aurora MySQL"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Storage Snapshot"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", Value: util.StringPtr("Aurora MySQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:BackupUsage$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB-Mo"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Backtrack",
				MonthlyQuantity: decimal.NewFromFloat(26006250),
				Unit:            "CR-Hr",
				Details:         []string{"Backtrack"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("System Operation"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:BacktrackUsage$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("CR-Hr"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Snapshot export",
				MonthlyQuantity: decimal.NewFromFloat(300),
				Unit:            "GB",
				Details:         []string{"Snapshot"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("System Operation"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", Value: util.StringPtr("Aurora MySQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:SnapshotExportToS3$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_rds_cluster")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

	t.Run("RDSClusterServerlessPostgres", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_rds_cluster.test",
			Type:         "aws_rds_cluster",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"engine":                  "aurora-postgresql",
				"backup_retention_period": 15,
				"storage_type":            "aurora-iopt1",
				"serverlessv2_scaling_configuration": []interface{}{
					map[string]interface{}{
						"min_capacity": 1,
					},
				},
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Storage (I/O-optimized)",
				MonthlyQuantity: decimal.NewFromFloat(50),
				Unit:            "GB-Mo",
				Details:         []string{"Storage (I/O-optimized)"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Database Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", ValueRegex: util.StringPtr("Aurora PostgreSQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:IO-OptimizedStorageUsage$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB-Mo"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "I/O requests",
				MonthlyQuantity: decimal.NewFromFloat(21024000),
				Unit:            "IOs",
				Details:         []string{"I/O requests"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("System Operation"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", ValueRegex: util.StringPtr("Aurora PostgreSQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:StorageIOUsage$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("IOs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Aurora ServerlessV2",
				MonthlyQuantity: decimal.NewFromFloat(0.5),
				Unit:            "ACU-Hr",
				Details:         []string{"Aurora PostgreSQL"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("ServerlessV2"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", Value: util.StringPtr("Aurora PostgreSQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:ServerlessV2IOOptimizedUsage$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("ACU-Hr"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
			{
				Name:            "Backup storage",
				MonthlyQuantity: decimal.NewFromFloat(840),
				Unit:            "GB-Mo",
				Details:         []string{"Aurora PostgreSQL"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("Storage Snapshot"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", Value: util.StringPtr("Aurora PostgreSQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:BackupUsage$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB-Mo"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},

			{
				Name:            "Snapshot export",
				MonthlyQuantity: decimal.NewFromFloat(300),
				Unit:            "GB",
				Details:         []string{"Snapshot"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonRDS"),
					Family:   util.StringPtr("System Operation"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "DatabaseEngine", Value: util.StringPtr("Aurora PostgreSQL")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*Aurora:SnapshotExportToS3$")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
						{Key: "StartingRange", Value: util.StringPtr("0")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_rds_cluster")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})
}
