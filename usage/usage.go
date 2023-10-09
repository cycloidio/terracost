package usage

const (
	// Key is the key used to set the usage
	// on the values passed to the resources
	Key string = "tc_usage"
)

// Default is the default Usage that will be used if none is configured
var Default = Usage{
	ResourceDefaultTypeUsage: map[string]interface{}{
		"aws_eks_node_group": map[string]interface{}{
			"instances":                        15,
			"operating_system":                 "linux",
			"reserved_instance_type":           "standard",
			"reserved_instance_term":           "1_year",
			"reserved_instance_payment_option": "partial_upfront",
			"monthly_cpu_credit_hrs":           350,
			"vcpu_count":                       2,
		},
		"aws_efs_file_system": map[string]interface{}{
			"storage_gb":                         180,
			"infrequent_access_storage_gb":       10,
			"monthly_infrequent_access_read_gb":  20,
			"monthly_infrequent_access_write_gb": 30,
		},
		"aws_fsx_openzfs_file_system": map[string]interface{}{
			"backup_storage_gb": 1024,
		},
		"aws_fsx_windows_file_system": map[string]interface{}{
			"backup_storage_gb": 1024,
		},
		"aws_fsx_ontap_file_system": map[string]interface{}{
			"backup_storage_gb": 1024,
		},
		"aws_fsx_lustre_file_system": map[string]interface{}{
			"backup_storage_gb": 1024,
		},
		"aws_nat_gateway": map[string]interface{}{
			"monthly_data_processed_gb": 10,
		},
	},
}

// Usage is the struct defining all the configure usages
type Usage struct {
	ResourceDefaultTypeUsage map[string]interface{} `json:"resource_default_type_usage",yaml:"resource_default_type_usage"`
}

// GetUsage will return the usage from the resource rt (ex: aws_instance)
func (u Usage) GetUsage(rt string) map[string]interface{} {
	us, ok := u.ResourceDefaultTypeUsage[rt]
	if ok {
		return us.(map[string]interface{})
	}

	return nil
}
