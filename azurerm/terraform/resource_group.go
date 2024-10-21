package terraform

import (
	"github.com/mitchellh/mapstructure"
)

// resourceGroupValues is holds the values that we need to be able
// to calculate the price of the ComputeInstance
type resourceGroupValues struct {
	Location string `mapstructure:"location"`
}

// decodeResourceGroupValues decodes and returns Values from a Terraform values map.
func decodeResourceGroupValues(tfVals map[string]interface{}) (resourceGroupValues, error) {
	var v resourceGroupValues
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
