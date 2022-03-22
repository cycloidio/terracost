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

func TestElastiCache_Components(t *testing.T) {
	p, err := NewProvider("aws", "us-east-1")
	require.NoError(t, err)

	t.Run("RedisEngine", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_cluster.test",
			Type:         "aws_elasticache_cluster",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type":       "cache.m4.large",
				"engine":          "redis",
				"num_cache_nodes": 1,
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

	t.Run("RedisSnapShotRetentionLimit", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_cluster.test",
			Type:         "aws_elasticache_cluster",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type":                "cache.m4.large",
				"engine":                   "redis",
				"num_cache_nodes":          1,
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

	t.Run("RedisReplicationGroupID", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_cluster.test",
			Type:         "aws_elasticache_cluster",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type":            "cache.m4.large",
				"engine":               "redis",
				"replication_group_id": "replication-group-1",
			},
		}

		expected := []query.Component{}

		actual := p.ResourceComponents(tfres)
		assert.Equal(t, expected, actual)
	})

	t.Run("MemcacheEngine", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_cluster.test",
			Type:         "aws_elasticache_cluster",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type":       "cache.m4.large",
				"engine":          "memcached",
				"num_cache_nodes": 1,
			},
		}

		expected := []query.Component{
			{
				Name:           "Cache instance",
				HourlyQuantity: decimal.NewFromInt(1),
				Details:        []string{"Memcached"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonElastiCache"),
					Family:   util.StringPtr("Cache Instance"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "InstanceType", Value: util.StringPtr("cache.m4.large")},
						{Key: "CacheEngine", Value: util.StringPtr("Memcached")},
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

	t.Run("MemcacheNumCacheNodes", func(t *testing.T) {
		tfres := terraform.Resource{
			Address:      "aws_elasticache_cluster.test",
			Type:         "aws_elasticache_cluster",
			Name:         "test",
			ProviderName: "aws",
			Values: map[string]interface{}{
				"node_type":       "cache.m4.large",
				"engine":          "memcached",
				"num_cache_nodes": 2,
			},
		}

		expected := []query.Component{
			{
				Name:           "Cache instance",
				HourlyQuantity: decimal.NewFromInt(2),
				Details:        []string{"Memcached"},
				ProductFilter: &product.Filter{
					Provider: util.StringPtr("aws"),
					Service:  util.StringPtr("AmazonElastiCache"),
					Family:   util.StringPtr("Cache Instance"),
					Location: util.StringPtr("us-east-1"),
					AttributeFilters: []*product.AttributeFilter{
						{Key: "InstanceType", Value: util.StringPtr("cache.m4.large")},
						{Key: "CacheEngine", Value: util.StringPtr("Memcached")},
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
