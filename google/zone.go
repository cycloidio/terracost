package google

import (
	"strings"
	"fmt"
)

// zoneToRegion will transform an europe-west1-b to europe-west1
func zoneToRegion(z string) (string, error) {
	spl := strings.Split(z, "-")
	if len(spl) != 3 {
		return "", fmt.Errorf("invalid zone: %s", z)
	}

	return strings.Join(spl[0:2], "-"), nil
}
