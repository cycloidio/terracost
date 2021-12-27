package terraform

import (
	"fmt"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
)

// Provider is an implementation of the terraform.Provider, used to extract component queries from
// terraform resources.
type Provider struct {
	key    string
	region region.Code
}

// NewProvider returns a new Provider with the provided default region and a query key.
func NewProvider(key string, regionCode region.Code) (*Provider, error) {
	if !regionCode.Valid() {
		return nil, fmt.Errorf("invalid AWS region: %q", regionCode)
	}
	return &Provider{key: key, region: regionCode}, nil
}

// Name returns the Provider's common name.
func (p *Provider) Name() string { return p.key }

// ResourceComponents returns Component queries for a given terraform.Resource.
func (p *Provider) ResourceComponents(tfRes terraform.Resource) []query.Component {
	switch tfRes.Type {
	case "aws_instance":
		vals, err := decodeInstanceValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newInstance(vals).Components()
	case "aws_ebs_volume":
		vals, err := decodeVolumeValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newVolume(vals).Components()
	case "aws_elasticache_cluster":
		vals, err := decodeElastiCacheValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newElastiCache(vals).Components()
	case "aws_db_instance":
		vals, err := decodeDBInstanceValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newDBInstance(vals).Components()
	case "aws_lb", "aws_alb":
		vals, err := decodeLBValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newLB(vals).Components()
	case "aws_elb":
		// ELB Classic does not have any special configuration.
		vals := lbValues{LoadBalancerType: "classic"}
		return p.newLB(vals).Components()
	default:
		return nil
	}
}
