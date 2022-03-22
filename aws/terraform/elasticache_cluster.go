package terraform

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/util"
)

// ElastiCache represents an ElastiCache instance definition that can be cost-estimated.
type ElastiCache struct {
	providerKey string

	region       region.Code
	instanceType string

	// cacheEngine can be one of "Memcached" or "Redis".
	cacheEngine string

	// # num_cache_nodes: The initial number of cache nodes that the cache cluster will have
	numCacheNodes decimal.Decimal

	// If replicationGroupID is set, aws_elasticache_replication_group will be use to define the cost
	replicationGroupID string

	snapshotRetentionLimit decimal.Decimal
}

type elastiCacheValues struct {
	NodeType               string `mapstructure:"node_type"`
	AvailabilityZone       string `mapstructure:"availability_zone"`
	Engine                 string `mapstructure:"engine"`
	ReplicationGroupID     string `mapstructure:"replication_group_id"`
	NumCacheNodes          int64  `mapstructure:"num_cache_nodes"`
	SnapshotRetentionLimit int64  `mapstructure:"snapshot_retention_limit"`
}

var cacheTypeMap = map[string]string{
	"memcached": "Memcached",
	"redis":     "Redis",
}

func decodeElastiCacheValues(tfVals map[string]interface{}) (elastiCacheValues, error) {
	var v elastiCacheValues
	if err := mapstructure.Decode(tfVals, &v); err != nil {
		return v, err
	}
	return v, nil
}

// NewInstance creates a new Instance from Terraform values.
func (p *Provider) newElastiCache(vals elastiCacheValues) *ElastiCache {
	cacheType := cacheTypeMap[vals.Engine]

	inst := &ElastiCache{
		providerKey:            p.key,
		region:                 p.region,
		instanceType:           vals.NodeType,
		cacheEngine:            cacheType,
		numCacheNodes:          decimal.NewFromInt(vals.NumCacheNodes),
		replicationGroupID:     vals.ReplicationGroupID,
		snapshotRetentionLimit: decimal.NewFromInt(vals.SnapshotRetentionLimit),
	}

	if reg := region.NewFromZone(vals.AvailabilityZone); reg.Valid() {
		inst.region = reg
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *ElastiCache) Components() []query.Component {
	// If replicationGroupID is set, aws_elasticache_replication_group will be use to define the cost
	if len(inst.replicationGroupID) > 0 {
		return []query.Component{}
	}

	components := []query.Component{inst.elastiCacheInstanceComponent()}

	if inst.snapshotRetentionLimit.GreaterThan(decimal.NewFromInt(0)) && strings.HasPrefix(inst.cacheEngine, "Redis") {
		components = append(components, inst.backupStorageComponent())
	}

	return components
}

func (inst *ElastiCache) elastiCacheInstanceComponent() query.Component {
	instClass := inst.cacheEngine
	attrFilters := []*product.AttributeFilter{
		{Key: "InstanceType", Value: util.StringPtr(inst.instanceType)},
		{Key: "CacheEngine", Value: util.StringPtr(inst.cacheEngine)},
	}

	return query.Component{
		Name:           "Cache instance",
		Details:        []string{instClass},
		HourlyQuantity: inst.numCacheNodes,
		ProductFilter: &product.Filter{
			Provider:         util.StringPtr(inst.providerKey),
			Service:          util.StringPtr("AmazonElastiCache"),
			Family:           util.StringPtr("Cache Instance"),
			Location:         util.StringPtr(inst.region.String()),
			AttributeFilters: attrFilters,
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("Hrs"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
			},
		},
	}
}

func (inst *ElastiCache) backupStorageComponent() query.Component {
	// MonthlyQuantity = snapshotRetentionLimit * backupSnapshotSize
	// TODO: If/When usage estimation will be supported, backupSnapshotSize might have a different value from 0

	backupSnapshotSize := decimal.NewFromInt(0)
	monthlyQuantityTotal := backupSnapshotSize.Mul(inst.snapshotRetentionLimit)

	return query.Component{
		Name:            "Backup storage",
		Details:         []string{monthlyQuantityTotal.String()},
		MonthlyQuantity: monthlyQuantityTotal,
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(inst.providerKey),
			Service:  util.StringPtr("AmazonElastiCache"),
			Family:   util.StringPtr("Storage Snapshot"),
			Location: util.StringPtr(inst.region.String()),
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("GB-Mo"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
			},
		},
	}
}
