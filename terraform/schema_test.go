package terraform_test

import (
	"encoding/json"
	"testing"

	"github.com/cycloidio/terracost/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderConfigUnmarshalJSON(t *testing.T) {
	raw := []byte(`
{
	"name": "azurerm",
	"expressions": {
		"client_id": {
			"references": [
				"var.azure_client_id"
			]
		},
		"features": [
			{}
		],
		"tenant_id": {
			"references": [
				"var.azure_tenant_id"
			]
		}
	}
}`)
	ex := terraform.ProviderConfig{
		Name: "azurerm",
		Expressions: map[string]terraform.ProviderConfigExpression{
			"client_id": terraform.ProviderConfigExpression{
				References: []string{"var.azure_client_id"},
			},
			"tenant_id": terraform.ProviderConfigExpression{
				References: []string{"var.azure_tenant_id"},
			},
		},
	}

	var pcfg terraform.ProviderConfig
	err := json.Unmarshal(raw, &pcfg)
	require.NoError(t, err)
	assert.Equal(t, ex, pcfg)
}
