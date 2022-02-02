package aws

// SupportedServices is a list of all AWS services that are supported by Terracost.
var supportedServices = map[string]struct{}{
	"AmazonEC2":         {},
	"AmazonRDS":         {},
	"AWSELB":            {},
	"AmazonElastiCache": {},
}

// IsServiceSupported returns true if the AWS service is valid and supported by Terracost (e.g. for ingestion.)
func IsServiceSupported(service string) bool {
	_, ok := supportedServices[service]
	return ok
}

// GetSupportedServices returns all the AWS service names that Terracost supports.
func GetSupportedServices() []string {
	svcs := make([]string, 0, len(supportedServices))
	for k := range supportedServices {
		svcs = append(svcs, k)
	}
	return svcs
}
