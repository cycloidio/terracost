package region_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cycloidio/terracost/aws/region"
)

func TestNewFromZone(t *testing.T) {
	testcases := []struct{ in, out string }{
		{"", ""},
		{"us-east-1c", "us-east-1"},
		{"eu-west-3a", "eu-west-3"},
	}
	for _, tc := range testcases {
		t.Run(tc.in, func(t *testing.T) {
			actual := region.NewFromZone(tc.in)
			assert.Equal(t, tc.out, actual.String())
		})
	}
}

func TestNewFromName(t *testing.T) {
	testcases := []struct{ in, out string }{
		{"", ""},
		{"US East (N. Virginia)", "us-east-1"},
		{"EU (Paris)", "eu-west-3"},
	}
	for _, tc := range testcases {
		t.Run(tc.in, func(t *testing.T) {
			actual := region.NewFromName(tc.in)
			assert.Equal(t, tc.out, actual.String())
		})
	}
}

func TestCode_Valid(t *testing.T) {
	testcases := []struct {
		in  string
		out bool
	}{
		{"", false},
		{"us-east-1", true},
		{"us-invalid-42", false},
	}
	for _, tc := range testcases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.out, region.Code(tc.in).Valid())
		})
	}
}
