package cost_test

import (
	"testing"

	"github.com/cycloidio/terracost/cost"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewHourly(t *testing.T) {
	val := decimal.NewFromFloat(1.23)
	c := cost.NewHourly(val)
	assertDecimalEqual(t, val.Mul(cost.HoursPerMonth), c.Decimal)
}

func TestNewMonthly(t *testing.T) {
	val := decimal.NewFromFloat(1.23)
	c := cost.NewMonthly(val)
	assertDecimalEqual(t, val, c.Decimal)
}

func TestCost_Hourly(t *testing.T) {
	val := decimal.NewFromFloat(1.23)
	c := cost.NewMonthly(val.Mul(cost.HoursPerMonth))
	assertDecimalEqual(t, val, c.Hourly())
}

func TestCost_Monthly(t *testing.T) {
	val := decimal.NewFromFloat(1.23)
	c := cost.NewHourly(val)
	assertDecimalEqual(t, val.Mul(cost.HoursPerMonth), c.Monthly())
}

func TestCost_Add(t *testing.T) {
	c1 := cost.NewMonthly(decimal.NewFromFloat(1.23))
	c2 := cost.NewMonthly(decimal.NewFromFloat(3.21))
	assertDecimalEqual(t, decimal.NewFromFloat(4.44), c1.Add(c2).Decimal)
}

func TestCost_MulDecimal(t *testing.T) {
	c := cost.NewMonthly(decimal.NewFromFloat(1.23))
	d := decimal.NewFromInt(3)
	assertDecimalEqual(t, decimal.NewFromFloat(3.69), c.MulDecimal(d).Decimal)
}

func assertDecimalEqual(t *testing.T, expected, actual decimal.Decimal) {
	assert.Truef(t, expected.Equal(actual), "Not equal:\nexpected: %s\nactual  : %s", expected, actual)
}
