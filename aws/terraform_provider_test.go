package aws_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/aws"
)

func TestTerraformProviderInitializer(t *testing.T) {

	initalizer := aws.TerraformProviderInitializer

	t.Run("WithoutRegion", func(t *testing.T) {
		p, err := initalizer.Provider(map[string]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, "aws", p.Name())
	})
	t.Run("WithRegion", func(t *testing.T) {
		p, err := initalizer.Provider(map[string]interface{}{"region": "eu-west-1"})
		require.NoError(t, err)
		assert.Equal(t, "aws", p.Name())
	})
	t.Run("WithEmptyRegion", func(t *testing.T) {
		p, err := initalizer.Provider(map[string]interface{}{"region": ""})
		// default region assigned -> no error
		require.NoError(t, err)
		assert.Equal(t, "aws", p.Name())
	})
	t.Run("WithNilRegion", func(t *testing.T) {
		_, err := initalizer.Provider(map[string]interface{}{"region": nil})
		require.Error(t, err)
	})
	t.Run("WithInvalidRegionType", func(t *testing.T) {
		_, err := initalizer.Provider(map[string]interface{}{"region": map[string]string{"foo": "bar"}})
		require.Error(t, err)
	})
}
