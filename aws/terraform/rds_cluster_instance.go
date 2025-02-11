package terraform

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"

	"github.com/cycloidio/terracost/aws/region"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/util"
)

// RDSClusterInstance represents an SQS queue definition that can be cost-estimated.
type RDSClusterInstance struct {
	provider                           *Provider
	region                             region.Code
	engine                             string
	engineVersion                      string
	instanceClass                      string
	storageType                        string
	performanceInsightsEnabled         bool
	performanceInsightsRetentionPeriod decimal.Decimal

	isServerless  bool
	isIOOptimized bool

	// Usage
	monthlyAdditionalPerformanceInsightsRequests decimal.Decimal
	capacityUnitsPerHr                           decimal.Decimal
}

type rdsClusterInstanceValues struct {
	ClusterIdentifier string `mapstructure:"cluster_identifier"`

	InstanceClass string `mapstructure:"instance_class"`
	StorageType   string `mapstructure:"storage_type"`

	Engine        string `mapstructure:"engine"`
	EngineVersion string `mapstructure:"engine_version"`

	PerformanceInsightsEnabled         bool    `mapstructure:"performance_insights_enabled"`
	PerformanceInsightsRetentionPeriod float64 `mapstructure:"performance_insights_retention_period"` // Long retention > 7

	Usage struct {
		MonthlyAdditionalPerformanceInsightsRequests float64 `mapstructure:"monthly_additional_performance_insights_requests"`
		CapacityUnitsPerHr                           float64 `mapstructure:"capacity_units_per_hr"`
	} `mapstructure:"tc_usage"`
}

// decodeRDSClusterInstanceValues decodes and returns rdsClusterInstanceValues from a Terraform values map.
func decodeRDSClusterInstanceValues(tfVals map[string]interface{}) (rdsClusterInstanceValues, error) {
	var v rdsClusterInstanceValues
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

// newRDSClusterInstance creates a new RDSClusterInstance from rdsClusterInstanceValues.
func (p *Provider) newRDSClusterInstance(_ map[string]terraform.Resource, vals rdsClusterInstanceValues) *RDSClusterInstance {
	v := &RDSClusterInstance{
		provider:                           p,
		region:                             p.region,
		isIOOptimized:                      false,
		instanceClass:                      vals.InstanceClass,
		performanceInsightsEnabled:         vals.PerformanceInsightsEnabled,
		performanceInsightsRetentionPeriod: decimal.NewFromFloat(vals.PerformanceInsightsRetentionPeriod),
		engine:                             vals.Engine,
		engineVersion:                      vals.EngineVersion,

		// Usage
		capacityUnitsPerHr:                           decimal.NewFromFloat(vals.Usage.CapacityUnitsPerHr),
		monthlyAdditionalPerformanceInsightsRequests: decimal.NewFromFloat(vals.Usage.MonthlyAdditionalPerformanceInsightsRequests),
	}

	v.isServerless = strings.EqualFold(vals.InstanceClass, "db.serverless")

	switch v.storageType {
	case "aurora-iopt1":
		v.isIOOptimized = true
	}

	return v
}

// Components returns the price component queries that make up the RDSClusterInstance.
func (v *RDSClusterInstance) Components() []query.Component {
	components := []query.Component{}

	databaseEngine := "Aurora MySQL"
	if v.engine == "aurora-postgresql" {
		databaseEngine = "Aurora PostgreSQL"
	}

	if v.isServerless {
		components = append(components, v.rdsClusterInstanceServerlessV2CostComponent(databaseEngine))
	} else {
		components = append(components, v.rdsClusterInstanceCostComponent(databaseEngine))
	}

	if v.performanceInsightsEnabled {

		if v.performanceInsightsRetentionPeriod.Cmp(decimal.NewFromInt(7)) > 0 {
			components = append(components, v.rdsClusterInstanceInsightsLongTermRetentionComponent(databaseEngine))
		}

		components = append(components, v.rdsClusterInstanceInsightsAPIRequestComponent())
	}
	return components
}

func (v *RDSClusterInstance) rdsClusterInstanceServerlessV2CostComponent(databaseEngine string) query.Component {

	name := "Aurora serverless v2"
	usageType := ".*Aurora:ServerlessV2Usage$"
	if v.isIOOptimized {
		name = "Aurora serverless v2 (I/O-optimized)"
		usageType = ".*Aurora:ServerlessV2IOOptimizedUsage$"
	}

	return query.Component{
		Name:           name,
		HourlyQuantity: v.capacityUnitsPerHr,
		Details:        []string{name},
		Usage:          true,
		Unit:           "ACU-Hr",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonRDS"),
			Family:   util.StringPtr("ServerlessV2"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "DatabaseEngine", Value: util.StringPtr(databaseEngine)},
				{Key: "UsageType", ValueRegex: util.StringPtr(usageType)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("ACU-Hr"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("0")},
			},
		},
	}
}

func (v *RDSClusterInstance) rdsClusterInstanceCostComponent(databaseEngine string) query.Component {

	usageType := ".*InstanceUsage:.*"
	if v.isIOOptimized {
		usageType = ".*InstanceUsageIOOptimized:.*"
	}

	return query.Component{
		Name:           "Database instance",
		HourlyQuantity: decimal.NewFromInt(1),
		Details:        []string{"instance"},
		Usage:          false,
		Unit:           "Hrs",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonRDS"),
			Family:   util.StringPtr("Database Instance"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "InstanceType", Value: util.StringPtr(v.instanceClass)},
				{Key: "DatabaseEngine", Value: util.StringPtr(databaseEngine)},
				{Key: "UsageType", ValueRegex: util.StringPtr(usageType)},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("Hrs"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("0")},
			},
		},
	}
}

func (v *RDSClusterInstance) rdsClusterInstanceInsightsLongTermRetentionComponent(databaseEngine string) query.Component {

	return query.Component{
		Name:            "Performance Insights Long Term Retention (serverless)",
		MonthlyQuantity: v.capacityUnitsPerHr,
		Details:         []string{"Insights"},
		Usage:           true,
		Unit:            "ACU-Months",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonRDS"),
			Family:   util.StringPtr("Performance Insights"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "DatabaseEngine", Value: util.StringPtr(databaseEngine)},
				{Key: "UsageType", ValueRegex: util.StringPtr(".*PI_LTR_FMR:Serverless$")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("ACU-Months"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("0")},
			},
		},
	}
}

func (v *RDSClusterInstance) rdsClusterInstanceInsightsAPIRequestComponent() query.Component {
	return query.Component{
		Name:            "Performance Insights API",
		MonthlyQuantity: v.monthlyAdditionalPerformanceInsightsRequests,
		Details:         []string{"Requests"},
		Usage:           true,
		Unit:            "API Calls",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonRDS"),
			Family:   util.StringPtr("Performance Insights"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "UsageType", ValueRegex: util.StringPtr(".*PI_API$")},
			},
		},
		PriceFilter: &price.Filter{
			Unit: util.StringPtr("API Calls"),
			AttributeFilters: []*price.AttributeFilter{
				{Key: "TermType", Value: util.StringPtr("OnDemand")},
				{Key: "StartingRange", Value: util.StringPtr("0")},
			},
		},
	}
}
