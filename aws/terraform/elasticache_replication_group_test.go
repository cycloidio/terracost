package terraform

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
)

func TestElastiCacheReplication_Components(t *testing.T) {
	p, err := NewProvider("aws", "us-east-1")
	require.NoError(t, err)

	//1 group 1 node
	t.Run("RedisEngineDefault", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_replication_group.test",
			Type:         "aws_elasticache_replication_group",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type":             "cache.m4.large",
				"number_cache_clusters": 1,
				"availability_zones":    []string{"us-east-1a", "us-east-1b"},
			},
		}

		expected := []query.Component{
			{
				Name:           "Cache instance",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Redis"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonElastiCache"),
					Family:   util.StringPtr("Cache Instance"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "InstanceType", Value: util.StringPtr("cache.m4.large")},
						{Key: "CacheEngine", Value: util.StringPtr("Redis")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("RedisEngineGlobalReplicationGroupID", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_replication_group.test",
			Type:         "aws_elasticache_replication_group",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"number_cache_clusters":       1,
				"availability_zones":          []string{"us-east-1a", "us-east-1b"},
				"global_replication_group_id": "global-replication-group-1",
			},
		}

		expected := []query.Component{}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("RedisSnapShotRetentionLimit", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_replication_group.test",
			Type:         "aws_elasticache_replication_group",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type":                "cache.m4.large",
				"engine":                   "redis",
				"number_cache_clusters":    1,
				"snapshot_retention_limit": 5,
			},
		}

		expected := []query.Component{
			{
				Name:           "Cache instance",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Redis"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonElastiCache"),
					Family:   util.StringPtr("Cache Instance"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "InstanceType", Value: util.StringPtr("cache.m4.large")},
						{Key: "CacheEngine", Value: util.StringPtr("Redis")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},
			{
				Name:            "Backup storage",
				Details:         []string{"0"},
				MonthlyQuantity: decimal.NewFromInt(0),
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonElastiCache"),
					Family:   util.StringPtr("Storage Snapshot"),
					Location: util.StringPtr("us-east-1"),
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("GB-Mo"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("RedisEngineNumCacheNodes", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_replication_group.test",
			Type:         "aws_elasticache_replication_group",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type":             "cache.m4.large",
				"engine":                "redis",
				"number_cache_clusters": 2,
			},
		}

		expected := []query.Component{
			{
				Name:           "Cache instance",
				HourlyQuantity: decimal.NewFromInt(2),
				Details:        []string{"Redis"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonElastiCache"),
					Family:   util.StringPtr("Cache Instance"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "InstanceType", Value: util.StringPtr("cache.m4.large")},
						{Key: "CacheEngine", Value: util.StringPtr("Redis")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("RedisEngineClusterMode", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_replication_group.test",
			Type:         "aws_elasticache_replication_group",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type": "cache.m4.large",
				"engine":    "redis",
				"cluster_mode": []map[string]int{
					map[string]int{
						"replicas_per_node_group": 3,
						"num_node_groups":         2,
					},
				},
			},
		}

		expected := []query.Component{
			{
				Name:           "Cache instance",
				HourlyQuantity: decimal.NewFromInt(8),
				Details:        []string{"Redis"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonElastiCache"),
					Family:   util.StringPtr("Cache Instance"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "InstanceType", Value: util.StringPtr("cache.m4.large")},
						{Key: "CacheEngine", Value: util.StringPtr("Redis")},
					},
				},
				PriceFilter: &price.Filter{
					Unit: util.StringPtr("Hrs"),
					AttributeFilters: []*price.AttributeFilter{
						{Key: "TermType", Value: util.StringPtr("OnDemand")},
					},
				},
			},
		}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

}
