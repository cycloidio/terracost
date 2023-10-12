package terraform_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	awstf "github.com/cycloidio/terracost/aws/terraform"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/testutil"
	"github.com/cycloidio/terracost/usage"
	"github.com/cycloidio/terracost/util"
)

func TestFSXXFileSystem_Components(t *testing.T) {
	p, err := awstf.NewProvider("aws", "eu-west-1")
	require.NoError(t, err)

	t.Run("LustreFileSystem", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_fsx_lustre_file_system.test",
			Type:         "aws_fsx_lustre_file_system",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"storage_capacity":                float64(1200),
				"deployment_type":                 "PERSISTENT_2",
				"per_unit_storage_throughput":     float64(125),
				"automatic_backup_retention_days": float64(10),
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Lustre Storage SSD",
				MonthlyQuantity: decimal.NewFromFloat(1200),
				Unit:            "GB-Mo",
				Details:         []string{"Storage", "Lustre"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("Persistent")},
						{Key: "FileSystemType", Value: util.StringPtr("Lustre")},
						{Key: "StorageType", Value: util.StringPtr("SSD")},
						{Key: "ThroughputCapacity", Value: util.StringPtr("125")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_fsx_lustre_file_system")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

	t.Run("OntapFileSystem", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_fsx_ontap_file_system.test",
			Type:         "aws_fsx_ontap_file_system",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"storage_capacity":                float64(1024),
				"deployment_type":                 "MULTI_AZ_1",
				"throughput_capacity":             float64(512),
				"automatic_backup_retention_days": float64(10),
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "ONTAP Storage SSD",
				MonthlyQuantity: decimal.NewFromFloat(1024),
				Unit:            "GB-Mo",
				Details:         []string{"Storage", "ONTAP"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("Multi-AZ")},
						{Key: "FileSystemType", Value: util.StringPtr("ONTAP")},
						{Key: "StorageType", Value: util.StringPtr("SSD")},
					},
				},
			},
			{
				Name:            "Throughput capacity",
				MonthlyQuantity: decimal.NewFromFloat(512),
				Unit:            "MiBps-Mo",
				Details:         []string{"Throughput capacity", "ONTAP"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Provisioned Throughput"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("Multi-AZ")},
						{Key: "FileSystemType", Value: util.StringPtr("ONTAP")},
					},
				},
			},
			{
				Name:            "ONTAP Backup storage",
				MonthlyQuantity: decimal.NewFromFloat(1024),
				Unit:            "GB-Mo",
				Details:         []string{"Storage", "ONTAP"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("N/A")},
						{Key: "FileSystemType", Value: util.StringPtr("ONTAP")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-BackupUsage")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_fsx_ontap_file_system")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

	t.Run("OpenzfsFileSystem", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_fsx_openzfs_file_system.test",
			Type:         "aws_fsx_openzfs_file_system",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"storage_capacity":                float64(1024),
				"deployment_type":                 "SINGLE_AZ_1",
				"throughput_capacity":             float64(64),
				"automatic_backup_retention_days": float64(10),
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "OpenZFS Storage SSD",
				MonthlyQuantity: decimal.NewFromFloat(1024),
				Unit:            "GB-Mo",
				Details:         []string{"Storage", "OpenZFS"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("Single-AZ")},
						{Key: "FileSystemType", Value: util.StringPtr("OpenZFS")},
						{Key: "StorageType", Value: util.StringPtr("SSD")},
					},
				},
			},
			{
				Name:            "Throughput capacity",
				MonthlyQuantity: decimal.NewFromFloat(64),
				Unit:            "MiBps-Mo",
				Details:         []string{"Throughput capacity", "OpenZFS"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Provisioned Throughput"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("Single-AZ")},
						{Key: "FileSystemType", Value: util.StringPtr("OpenZFS")},
					},
				},
			},
			{
				Name:            "OpenZFS Backup storage",
				MonthlyQuantity: decimal.NewFromFloat(1024),
				Unit:            "GB-Mo",
				Details:         []string{"Storage", "OpenZFS"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("N/A")},
						{Key: "FileSystemType", Value: util.StringPtr("OpenZFS")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-BackupUsage")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_fsx_openzfs_file_system")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

	t.Run("WindowsFileSystem", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_fsx_windows_file_system.test",
			Type:         "aws_fsx_windows_file_system",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"storage_capacity":                float64(300),
				"deployment_type":                 "MULTI_AZ_1",
				"throughput_capacity":             float64(1024),
				"automatic_backup_retention_days": float64(10),
			},
		}
		rss := map[string]terraform.Resource{}

		expected := []query.Component{
			{
				Name:            "Windows Storage SSD",
				MonthlyQuantity: decimal.NewFromFloat(300),
				Unit:            "GB-Mo",
				Details:         []string{"Storage", "Windows"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("Multi-AZ")},
						{Key: "FileSystemType", Value: util.StringPtr("Windows")},
						{Key: "StorageType", Value: util.StringPtr("SSD")},
					},
				},
			},
			{
				Name:            "Throughput capacity",
				MonthlyQuantity: decimal.NewFromFloat(1024),
				Unit:            "MiBps-Mo",
				Details:         []string{"Throughput capacity", "Windows"},
				Usage:           false,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Provisioned Throughput"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("Multi-AZ")},
						{Key: "FileSystemType", Value: util.StringPtr("Windows")},
					},
				},
			},
			{
				Name:            "Windows Backup storage",
				MonthlyQuantity: decimal.NewFromFloat(300),
				Unit:            "GB-Mo",
				Details:         []string{"Storage", "Windows"},
				Usage:           true,
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonFSx"),
					Family:   util.StringPtr("Storage"),
					Location: util.StringPtr("eu-west-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "Deployment_option", Value: util.StringPtr("Multi-AZ")},
						{Key: "FileSystemType", Value: util.StringPtr("Windows")},
						{Key: "UsageType", ValueRegex: util.StringPtr(".*-BackupUsage")},
					},
				},
			},
		}

		us := usage.Default.GetUsage("aws_fsx_windows_file_system")
		tfres.Values[usage.Key] = us
		actual := p.ResourceComponents(rss, tfres)
		testutil.EqualQueryComponents(t, expected, actual)
	})

}
