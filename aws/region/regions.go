package region

// List of regions with their codes can be found here: https://docs.aws.amazon.com/general/latest/gr/ec2-service.html
var nameToCode = map[string]Code{
	"US East (N. Virginia)":      "us-east-1",
	"US East (Ohio)":             "us-east-2",
	"US West (N. California)":    "us-west-1",
	"US West (Oregon)":           "us-west-2",
	"US West (Los Angeles)":      "us-west-2-lax-1", // Included in the offer file but missing from the docs.
	"Canada (Central)":           "ca-central-1",
	"Canada West (Calgary)":      "ca-west-1",
	"EU (Stockholm)":             "eu-north-1",
	"EU (Ireland)":               "eu-west-1",
	"EU (London)":                "eu-west-2",
	"EU (Paris)":                 "eu-west-3",
	"EU (Frankfurt)":             "eu-central-1",
	"EU (Zurich)":                "eu-central-2",
	"EU (Milan)":                 "eu-south-1",
	"EU (Spain)":                 "eu-south-2",
	"Asia Pacific (Tokyo)":       "ap-northeast-1",
	"Asia Pacific (Seoul)":       "ap-northeast-2",
	"Asia Pacific (Osaka-Local)": "ap-northeast-3",
	"Asia Pacific (Singapore)":   "ap-southeast-1",
	"Asia Pacific (Sydney)":      "ap-southeast-2",
	"Asie Pacific (Jakarta)":     "ap-southeast-3",
	"Asia Pacific (Melbourne)":   "ap-southeast-4",
	"Asie Pacific (Malaysia)":    "ap-southeast-5",
	"Asie Pacific (Thailand)":    "ap-southeast-7",
	"Asia Pacific (Hong Kong)":   "ap-east-1",
	"Asia Pacific (Mumbai)":      "ap-south-1",
	"Asie Pacific (Hyderabad)":   "ap-south-2",
	"South America (Sao Paulo)":  "sa-east-1",
	"China (Beijing)":            "cn-north-1",
	"China (Ningxia)":            "cn-northwest-1",
	"AWS GovCloud (US-West)":     "us-gov-west-1",
	"AWS GovCloud (US-East)":     "us-gov-east-1",
	"Middle East (Bahrain)":      "me-south-1",
	"Africa (Cape Town)":         "af-south-1",
	"Israel (Tel Aviv)":          "il-central-1",
	"Middle East (UAE)":          "me-central-1",
}

var codeToShortName = map[string]string{
	"us-east-1":       "",
	"us-east-2":       "USE2",
	"us-west-1":       "USW1",
	"us-west-2":       "USW2",
	"us-west-2-lax-1": "LAX1",
	"ca-central-1":    "CAN1",
	"ca-west-1":       "CAW1",
	"eu-north-1":      "EUN1",
	"eu-west-1":       "EU",
	"eu-west-2":       "EUW2",
	"eu-west-3":       "EUW3",
	"eu-central-1":    "EUC1",
	"eu-central-2":    "EUC2",
	"eu-south-1":      "EUS1",
	"eu-south-2":      "EUS2",
	"ap-south-1":      "APS3",
	"ap-northeast-1":  "APN1",
	"ap-northeast-2":  "APN2",
	"ap-northeast-3":  "APN3",
	"ap-southeast-1":  "APS1",
	"ap-southeast-2":  "APS2",
	"ap-southeast-3":  "APS3",
	"ap-southeast-4":  "APS6",
	"ap-southeast-5":  "APS7",
	"ap-southeast-7":  "APS9",
	"ap-east-1":       "APE1",
	"ap-south-2":      "APS5",
	"sa-east-1":       "SAE1",
	"cn-north-1":      "", // error
	"cn-northwest-1":  "", // error
	"us-gov-west-1":   "UGW1",
	"us-gov-east-1":   "UGE1",
	"me-south-1":      "MES1",
	"af-south-1":      "AFS1",
	"il-central-1":    "ILC1",
	"me-central-1":    "MEC1",
}

var codeToName = make(map[Code]string)

func init() {
	for name, code := range nameToCode {
		codeToName[code] = name
	}
}

func GetRegionToShortName(region string) string {
	return codeToShortName[region]
}
