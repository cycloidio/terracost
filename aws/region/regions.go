package region

// List of regions with their codes can be found here: https://docs.aws.amazon.com/general/latest/gr/ec2-service.html
var nameToCode = map[string]Code{
	"US East (N. Virginia)":      "us-east-1",
	"US East (Ohio)":             "us-east-2",
	"US West (N. California)":    "us-west-1",
	"US West (Oregon)":           "us-west-2",
	"US West (Los Angeles)":      "us-west-2-lax-1", // Included in the offer file but missing from the docs.
	"Canada (Central)":           "ca-central-1",
	"EU (Stockholm)":             "eu-north-1",
	"EU (Ireland)":               "eu-west-1",
	"EU (London)":                "eu-west-2",
	"EU (Paris)":                 "eu-west-3",
	"EU (Frankfurt)":             "eu-central-1",
	"EU (Milan)":                 "eu-south-1",
	"Asia Pacific (Mumbai)":      "ap-south-1",
	"Asia Pacific (Tokyo)":       "ap-northeast-1",
	"Asia Pacific (Seoul)":       "ap-northeast-2",
	"Asia Pacific (Osaka-Local)": "ap-northeast-3",
	"Asia Pacific (Singapore)":   "ap-southeast-1",
	"Asia Pacific (Sydney)":      "ap-southeast-2",
	"Asia Pacific (Hong Kong)":   "ap-east-1",
	"South America (Sao Paulo)":  "sa-east-1",
	"China (Beijing)":            "cn-north-1",
	"China (Ningxia)":            "cn-northwest-1",
	"AWS GovCloud (US-West)":     "us-gov-west-1",
	"AWS GovCloud (US-East)":     "us-gov-east-1",
	"Middle East (Bahrain)":      "me-south-1",
	"Africa (Cape Town)":         "af-south-1",
}

var codeToName = make(map[Code]string)

func init() {
	for name, code := range nameToCode {
		codeToName[code] = name
	}
}
