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

// RDSCluster represents an SQS queue definition that can be cost-estimated.
type RDSCluster struct {
	provider              *Provider
	region                region.Code
	engineMode            string
	engine                string
	storageType           string
	backupRetentionPeriod decimal.Decimal

	isServerless bool

	// Usage
	writeRequestsPerSec       decimal.Decimal
	readRequestsPerSec        decimal.Decimal
	changeRecordsPerStatement decimal.Decimal
	storageGB                 decimal.Decimal
	averageStatementsPerHr    decimal.Decimal
	backtrackWindowHrs        decimal.Decimal
	snapshotExportSizeGB      decimal.Decimal
	capacityUnitsPerHr        decimal.Decimal
	backupSnapshotSizeGB      decimal.Decimal
}

type rdsClusterValues struct {
	EngineMode                       string  `mapstructure:"engine_mode"`
	Engine                           string  `mapstructure:"engine"`
	StorageType                      string  `mapstructure:"storage_type"`
	BackupRetentionPeriod            float64 `mapstructure:"backup_retention_period"`
	Serverlessv2ScalingConfiguration []struct {
		MinCapacity float64 `mapstructure:"min_capacity"`
	} `mapstructure:"serverlessv2_scaling_configuration"`

	Usage struct {
		WriteRequestsPerSec       float64 `mapstructure:"write_requests_per_sec"`
		ReadRequestsPerSec        float64 `mapstructure:"read_requests_per_sec"`
		ChangeRecordsPerStatement float64 `mapstructure:"change_records_per_statement"`
		StorageGB                 float64 `mapstructure:"storage_gb"`
		AverageStatementsPerHr    float64 `mapstructure:"average_statements_per_hr"`
		BacktrackWindowHrs        float64 `mapstructure:"backtrack_window_hrs"`
		SnapshotExportSizeGB      float64 `mapstructure:"snapshot_export_size_gb"`
		CapacityUnitsPerHr        float64 `mapstructure:"capacity_units_per_hr"`
		BackupSnapshotSizeGB      float64 `mapstructure:"backup_snapshot_size_gb"`
	} `mapstructure:"tc_usage"`
}

// decodeRDSClusterValues decodes and returns rdsClusterValues from a Terraform values map.
func decodeRDSClusterValues(tfVals map[string]interface{}) (rdsClusterValues, error) {
	var v rdsClusterValues
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

// newRDSCluster creates a new RDSCluster from rdsClusterValues.
func (p *Provider) newRDSCluster(rss map[string]terraform.Resource, vals rdsClusterValues) *RDSCluster {
	// The 'rss' variable contains information from linked resources.
	// Currently, it is not utilized in this resource.
	_ = rss

	v := &RDSCluster{
		provider:              p,
		region:                p.region,
		engine:                vals.Engine,
		engineMode:            "provisioned",
		storageType:           vals.StorageType,
		backupRetentionPeriod: decimal.NewFromFloat(1),

		// Usage
		writeRequestsPerSec:       decimal.NewFromFloat(vals.Usage.WriteRequestsPerSec),
		readRequestsPerSec:        decimal.NewFromFloat(vals.Usage.ReadRequestsPerSec),
		changeRecordsPerStatement: decimal.NewFromFloat(vals.Usage.ChangeRecordsPerStatement),
		storageGB:                 decimal.NewFromFloat(vals.Usage.StorageGB),
		averageStatementsPerHr:    decimal.NewFromFloat(vals.Usage.AverageStatementsPerHr),
		backtrackWindowHrs:        decimal.NewFromFloat(vals.Usage.BacktrackWindowHrs),
		snapshotExportSizeGB:      decimal.NewFromFloat(vals.Usage.SnapshotExportSizeGB),
		capacityUnitsPerHr:        decimal.NewFromFloat(vals.Usage.CapacityUnitsPerHr),
		backupSnapshotSizeGB:      decimal.NewFromFloat(vals.Usage.BackupSnapshotSizeGB),
	}

	if vals.BackupRetentionPeriod > 1 {
		v.backupRetentionPeriod = decimal.NewFromFloat(vals.BackupRetentionPeriod)
	}

	if vals.EngineMode != "" {
		v.engineMode = vals.EngineMode
	}

	if len(vals.Serverlessv2ScalingConfiguration) > 0 || v.engineMode == "serverless" {
		v.isServerless = true
	}

	return v
}

// Components returns the price component queries that make up the RDSCluster.
func (v *RDSCluster) Components() []query.Component {

	isIOOptimized := false
	switch v.storageType {
	case "aurora-iopt1":
		isIOOptimized = true
	}

	databaseEngine := "Aurora MySQL"
	switch v.engine {
	case "aurora", "aurora-mysql":
		databaseEngine = "Aurora MySQL"
	case "aurora-postgresql":
		databaseEngine = "Aurora PostgreSQL"
	}

	components := v.rdsClusterAuroraStorageComponent(databaseEngine, isIOOptimized)

	if v.isServerless {
		components = append(components, v.rdsClusterAuroraServerlessComponent(databaseEngine))
	}

	if v.backupRetentionPeriod.GreaterThan(decimal.NewFromFloat(1)) {

		totalBackupStorageGB := v.backupSnapshotSizeGB.Mul(v.backupRetentionPeriod).Sub(v.backupSnapshotSizeGB)
		components = append(components, v.rdsClusterAuroraBackupComponent(totalBackupStorageGB, databaseEngine))
	}

	if !v.isServerless && !strings.Contains(v.engine, "postgresql") {
		totalBacktrackChangeRecords := v.averageStatementsPerHr.Mul(decimal.NewFromInt(730)).Mul(v.changeRecordsPerStatement).Mul(v.backtrackWindowHrs)
		components = append(components, v.rdsClusterAuroraBacktrackComponent(totalBacktrackChangeRecords))
	}

	components = append(components, v.rdsClusterAuroraSnapshotExportComponent(databaseEngine))
	return components
}

func (v *RDSCluster) rdsClusterAuroraServerlessComponent(databaseEngine string) query.Component {
	return query.Component{
		Name:            "Aurora serverless",
		MonthlyQuantity: v.capacityUnitsPerHr,
		Details:         []string{databaseEngine},
		Usage:           true,
		Unit:            "ACU-Hr",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonRDS"),
			Family:   util.StringPtr("Serverless"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "DatabaseEngine", Value: util.StringPtr(databaseEngine)},
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

func (v *RDSCluster) rdsClusterAuroraStorageComponent(databaseEngine string, isIOOptimized bool) []query.Component {

	name := "Storage"
	usageType := ".*Aurora:StorageUsage$"
	requestDatabaseEngineStorageType := databaseEngine

	if isIOOptimized {
		name = "Storage (I/O-optimized)"
		usageType = ".*Aurora:IO-OptimizedStorageUsage$"
	} else {
		if databaseEngine != "Aurora PostgreSQL" {
			databaseEngine = "Any"
		}
	}

	if databaseEngine != "Aurora PostgreSQL" {
		requestDatabaseEngineStorageType = "Any"
	}

	ioPerSecond := v.writeRequestsPerSec.Add(v.readRequestsPerSec)
	monthlyIORequests := ioPerSecond.Mul(decimal.NewFromInt(730)).Mul(decimal.NewFromInt(60)).Mul(decimal.NewFromInt(60))
	return []query.Component{
		{
			Name:            name,
			MonthlyQuantity: v.storageGB,
			Details:         []string{name},
			Usage:           true,
			Unit:            "GB-Mo",
			ProductFilter: &product.Filter{
				Provider: util.StringPtr(v.provider.key),
				Service:  util.StringPtr("AmazonRDS"),
				Family:   util.StringPtr("Database Storage"),
				Location: util.StringPtr(v.region.String()),
				AttributeFilters: []*product.AttributeFilter{
					{Key: "DatabaseEngine", ValueRegex: util.StringPtr(databaseEngine)},
					{Key: "UsageType", ValueRegex: util.StringPtr(usageType)},
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
			MonthlyQuantity: monthlyIORequests,
			Details:         []string{"I/O requests"},
			Usage:           true,
			Unit:            "IOs",
			ProductFilter: &product.Filter{
				Provider: util.StringPtr(v.provider.key),
				Service:  util.StringPtr("AmazonRDS"),
				Family:   util.StringPtr("System Operation"),
				Location: util.StringPtr(v.region.String()),
				AttributeFilters: []*product.AttributeFilter{
					{Key: "DatabaseEngine", ValueRegex: util.StringPtr(requestDatabaseEngineStorageType)},
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
	}
}

func (v *RDSCluster) rdsClusterAuroraBackupComponent(totalBackupStorageGB decimal.Decimal, databaseEngine string) query.Component {

	return query.Component{
		Name:            "Backup storage",
		MonthlyQuantity: totalBackupStorageGB,
		Details:         []string{databaseEngine},
		Usage:           true,
		Unit:            "GB-Mo",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonRDS"),
			Family:   util.StringPtr("Storage Snapshot"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "DatabaseEngine", Value: util.StringPtr(databaseEngine)},
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
	}
}

func (v *RDSCluster) rdsClusterAuroraBacktrackComponent(totalBacktrackChangeRecords decimal.Decimal) query.Component {
	return query.Component{
		Name:            "Backtrack",
		MonthlyQuantity: totalBacktrackChangeRecords,
		Details:         []string{"Backtrack"},
		Usage:           true,
		Unit:            "CR-Hr",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonRDS"),
			Family:   util.StringPtr("System Operation"),
			Location: util.StringPtr(v.region.String()),
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
	}
}

func (v *RDSCluster) rdsClusterAuroraSnapshotExportComponent(databaseEngine string) query.Component {
	return query.Component{
		Name:            "Snapshot export",
		MonthlyQuantity: v.snapshotExportSizeGB,
		Details:         []string{"Snapshot"},
		Usage:           true,
		Unit:            "GB",
		ProductFilter: &product.Filter{
			Provider: util.StringPtr(v.provider.key),
			Service:  util.StringPtr("AmazonRDS"),
			Family:   util.StringPtr("System Operation"),
			Location: util.StringPtr(v.region.String()),
			AttributeFilters: []*product.AttributeFilter{
				{Key: "DatabaseEngine", Value: util.StringPtr(databaseEngine)},
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
	}
}
