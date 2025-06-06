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
func (p *Provider) ResourceComponents(rss map[string]terraform.Resource, tfRes terraform.Resource) []query.Component {
	switch tfRes.Type {
	case "aws_instance":
		vals, err := decodeInstanceValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newInstance(vals).Components()
	case "aws_autoscaling_group":
		vals, err := decodeAutoscalingGroupValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newAutoscalingGroup(rss, vals).Components()
	case "aws_cloudwatch_log_group":
		vals, err := decodeCloudwatchLogGroupValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newCloudwatchLogGroup(rss, vals).Components()
	case "aws_cloudwatch_metric_alarm":
		vals, err := decodeCloudwatchMetricAlarmValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newCloudwatchMetricAlarm(rss, vals).Components()

	case "aws_db_instance":
		vals, err := decodeDBInstanceValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newDBInstance(vals).Components()
	case "aws_ebs_volume":
		vals, err := decodeVolumeValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newVolume(vals).Components()
	case "aws_efs_file_system":
		vals, err := decodeEFSFileSystemValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newEFSFileSystem(rss, vals).Components()
	case "aws_elasticache_cluster":
		vals, err := decodeElastiCacheValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newElastiCache(vals).Components()
	case "aws_elasticache_replication_group":
		vals, err := decodeElastiCacheReplicationValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newElastiCacheReplication(vals).Components()
	case "aws_eip":
		vals, err := decodeElasticIPValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newElasticIP(vals).Components()
	case "aws_elb":
		// ELB Classic does not have any special configuration.
		vals := lbValues{LoadBalancerType: "classic"}
		return p.newLB(vals).Components()
	case "aws_eks_cluster":
		vals, err := decodeEKSClusterValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newEKSCluster(vals).Components()
	case "aws_eks_node_group":
		vals, err := decodeEKSNodeGroupValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newEKSNodeGroup(rss, vals).Components()
	case "aws_fsx_lustre_file_system":
		vals, err := decodeFSxLustreFileSystemValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newFSxLustreFileSystem(rss, vals).Components()
	case "aws_fsx_ontap_file_system":
		vals, err := decodeFSxOntapFileSystemValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newFSxOntapFileSystem(rss, vals).Components()
	case "aws_fsx_openzfs_file_system":
		vals, err := decodeFSxOpenzfsFileSystemValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newFSxOpenzfsFileSystem(rss, vals).Components()
	case "aws_fsx_windows_file_system":
		vals, err := decodeFSxWindowsFileSystemValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newFSxWindowsFileSystem(rss, vals).Components()
	case "aws_kms_key":
		vals, err := decodeKMSKeyValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newKMSKey(rss, vals).Components()
	case "aws_lb", "aws_alb":
		vals, err := decodeLBValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newLB(vals).Components()
	case "aws_nat_gateway":
		vals, err := decodeNatGatewayValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newNatGateway(vals).Components()
	case "aws_rds_cluster":
		vals, err := decodeRDSClusterValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newRDSCluster(rss, vals).Components()
	case "aws_rds_cluster_instance":
		vals, err := decodeRDSClusterInstanceValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newRDSClusterInstance(rss, vals).Components()
	case "aws_s3_bucket":
		vals, err := decodeS3BucketValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newS3Bucket(rss, vals).Components()
	case "aws_s3_bucket_analytics_configuration":
		vals, err := decodeS3BucketAnalyticsConfigurationValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newS3BucketAnalyticsConfiguration(rss, vals).Components()
	case "aws_s3_bucket_inventory":
		vals, err := decodeS3BucketInventoryValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newS3BucketInventory(rss, vals).Components()
	case "aws_secretsmanager_secret":
		vals, err := decodeSecretsmanagerSecretValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newSecretsmanagerSecret(rss, vals).Components()
	case "aws_sqs_queue":
		vals, err := decodeSQSQueueValues(tfRes.Values)
		if err != nil {
			return nil
		}
		return p.newSQSQueue(rss, vals).Components()
	default:
		return nil
	}
}
