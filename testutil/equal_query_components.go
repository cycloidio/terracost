package testutil

import (
	"fmt"
	"testing"

	"github.com/cycloidio/terracost/query"
	"github.com/stretchr/testify/assert"
)

// EqualQueryComponents will compare the components but the MonthlyQuantity will be
// compared via String and the rest with assert.Equal
func EqualQueryComponents(t *testing.T, eqcs, aqcs []query.Component) {
	t.Helper()

	for i, eqc := range eqcs {
		if eqc.MonthlyQuantity.String() == aqcs[i].MonthlyQuantity.String() {
			eqc.MonthlyQuantity = aqcs[i].MonthlyQuantity
		} else {
			assert.Fail(t, fmt.Sprintf("Expected MonthlyQuantity to be %q but was %q", eqc.MonthlyQuantity.String(), aqcs[i].MonthlyQuantity.String()))
			continue
		}
		assert.Equal(t, eqc, aqcs[i])
	}
}
