package google

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZoneToRegion(t *testing.T) {
	type args struct {
		z string
	}
	tests := []struct {
		name string
		args args
		want string
		err  error
	}{
		{
			name: "Success",
			args: args{
				z: "europe-west1-b",
			},
			want: "europe-west1",
			err:  nil,
		},
		{
			name: "FailInvalidZone",
			args: args{
				z: "zone",
			},
			want: "",
			err:  fmt.Errorf("invalid zone: %s", "zone"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := zoneToRegion(tt.args.z)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error(), "zoneToRegion(%v)", tt.args.z)
			}
			assert.Equalf(t, tt.want, got, "zoneToRegion(%v)", tt.args.z)
		})
	}
}
