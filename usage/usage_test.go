package usage_test

import (
	"testing"

	"github.com/cycloidio/terracost/usage"
	"github.com/stretchr/testify/assert"
)

func TestGetUsage(t *testing.T) {
	eu := map[string]interface{}{
		"something": "else",
	}
	us := usage.Usage{
		ResourceDefaultTypeUsage: map[string]interface{}{
			"aws_instance": eu,
		},
	}

	ru := us.GetUsage("aws_instance")
	assert.Equal(t, eu, ru)

}
