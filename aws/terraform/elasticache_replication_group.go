package terraform

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/query"
)

// ElastiCacheReplication represents an ElastiCacheReplication instance definition that can be cost-estimated.
type ElastiCacheReplication struct {
	providerKey string

	region       region.Code
	instanceType string

	// cacheEngine can be one of "Memcached" or "Redis".
	cacheEngine string

	// # num_cache_nodes: The initial number of cache nodes that the cache cluster will have
	numCacheNodes decimal.Decimal

	snapshotRetentionLimit decimal.Decimal

	globalReplicationGroupID string
}

type elastiCacheReplicationValues struct {
	NodeType             string   `mapstructure:"node_type"`
	AvailabilityZones    []string `mapstructure:"availability_zones"`
	Engine               string   `mapstructure:"engine"`
	NumNodeGroups        int64    `mapstructure:"num_node_groups"`
	ReplicasPerNodeGroup int64    `mapstructure:"replicas_per_node_group"`
	// Deprecated params cluster_mode
	ClusterMode []struct {
		NumNodeGroups        int64 `mapstructure:"num_node_groups"`
		ReplicasPerNodeGroup int64 `mapstructure:"replicas_per_node_group"`
	} `mapstructure:"cluster_mode"`
	NumberCacheClusters      int64  `mapstructure:"num_cache_clusters"`
	SnapshotRetentionLimit   int64  `mapstructure:"snapshot_retention_limit"`
	GlobalReplicationGroupID string `mapstructure:"global_replication_group_id"`
}

func decodeElastiCacheReplicationValues(tfVals map[string]interface{}) (elastiCacheReplicationValues, error) {
	var v elastiCacheReplicationValues
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &v,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return v, err
	}

	if err := decoder.Decode(tfVals); err != nil {
		return v, err
	}
	return v, nil
}

// NewInstance creates a new Instance from Terraform values.
func (p *Provider) newElastiCacheReplication(vals elastiCacheReplicationValues) *ElastiCacheReplication {

	cacheType := "Redis"
	if len(vals.Engine) > 0 {
		// cacheTypeMap from elasticache_cluster.go
		cacheType = cacheTypeMap[vals.Engine]
	}

	numCacheNodes := decimal.NewFromInt(vals.NumberCacheClusters)
	if len(vals.ClusterMode) > 0 {
		nodeGroups := decimal.NewFromInt(vals.ClusterMode[0].NumNodeGroups)
		replicasNode := decimal.NewFromInt(vals.ClusterMode[0].ReplicasPerNodeGroup)
		numCacheNodes = nodeGroups.Mul(replicasNode).Add(nodeGroups)
	} else if vals.NumNodeGroups > 0 {
		nodeGroups := decimal.NewFromInt(vals.NumNodeGroups)
		replicasNode := decimal.NewFromInt(vals.ReplicasPerNodeGroup)
		numCacheNodes = nodeGroups.Mul(replicasNode).Add(nodeGroups)
	}

	inst := &ElastiCacheReplication{
		providerKey:              p.key,
		region:                   p.region,
		instanceType:             vals.NodeType,
		cacheEngine:              cacheType,
		numCacheNodes:            numCacheNodes,
		snapshotRetentionLimit:   decimal.NewFromInt(vals.SnapshotRetentionLimit),
		globalReplicationGroupID: vals.GlobalReplicationGroupID,
	}

	if len(vals.AvailabilityZones) > 0 {
		if reg := region.NewFromZone(vals.AvailabilityZones[0]); reg.Valid() {
			inst.region = reg
		}
	}

	return inst
}

// Components returns the price component queries that make up this Instance.
func (inst *ElastiCacheReplication) Components() []query.Component {
	// If global_replication_group_id is set, node_type & num_node_groups can't be defined. So no cost found
	if len(inst.globalReplicationGroupID) > 0 {
		return []query.Component{}
	}

	components := []query.Component{inst.elastiCacheReplicationInstanceComponent()}

	if inst.snapshotRetentionLimit.GreaterThan(decimal.NewFromInt(0)) && strings.HasPrefix(inst.cacheEngine, "Redis") {
		components = append(components, inst.backupStorageComponent())
	}

	return components
}

func (inst *ElastiCacheReplication) elastiCacheReplicationInstanceComponent() query.Component {

	// Currently, cost is the same as ElastiCache
	// Use ElastiCache function to generate the right query
	elastiCacheInst := &ElastiCache{
		providerKey:            inst.providerKey,
		region:                 inst.region,
		instanceType:           inst.instanceType,
		cacheEngine:            inst.cacheEngine,
		numCacheNodes:          inst.numCacheNodes,
		snapshotRetentionLimit: inst.snapshotRetentionLimit,
	}

	return elastiCacheInst.elastiCacheInstanceComponent()

}

func (inst *ElastiCacheReplication) backupStorageComponent() query.Component {

	// Currently, cost is the same as ElastiCache
	// Use ElastiCache function to generate the right query
	elastiCacheInst := &ElastiCache{
		providerKey:            inst.providerKey,
		region:                 inst.region,
		instanceType:           inst.instanceType,
		cacheEngine:            inst.cacheEngine,
		numCacheNodes:          inst.numCacheNodes,
		snapshotRetentionLimit: inst.snapshotRetentionLimit,
	}

	return elastiCacheInst.backupStorageComponent()
}
