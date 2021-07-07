package cost_test

import (
	"testing"

	"github.com/cycloidio/terracost/cost"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewHourly(t *testing.T) {
	val := decimal.NewFromFloat(1.23)
	c := cost.NewHourly(val, "USD")
	assertDecimalEqual(t, val.Mul(cost.HoursPerMonth), c.Decimal)
}

func TestNewMonthly(t *testing.T) {
	val := decimal.NewFromFloat(1.23)
	c := cost.NewMonthly(val, "USD")
	assertDecimalEqual(t, val, c.Decimal)
}

func TestCost_Hourly(t *testing.T) {
	val := decimal.NewFromFloat(1.23)
	c := cost.NewMonthly(val.Mul(cost.HoursPerMonth), "USD")
	assertDecimalEqual(t, val, c.Hourly())
}

func TestCost_Monthly(t *testing.T) {
	val := decimal.NewFromFloat(1.23)
	c := cost.NewHourly(val, "USD")
	assertDecimalEqual(t, val.Mul(cost.HoursPerMonth), c.Monthly())
}

func TestCost_Add(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c1 := cost.NewMonthly(decimal.NewFromFloat(1.23), "USD")
		c2 := cost.NewMonthly(decimal.NewFromFloat(3.21), "USD")
		ac, err := c1.Add(c2)
		assert.NoError(t, err)
		assertDecimalEqual(t, decimal.NewFromFloat(4.44), ac.Decimal)
	})
	t.Run("CurrencyMismatch", func(t *testing.T) {
		c1 := cost.NewMonthly(decimal.NewFromFloat(1.23), "USD")
		c2 := cost.NewMonthly(decimal.NewFromFloat(3.21), "EUR")
		_, err := c1.Add(c2)
		assert.EqualError(t, err, "currency mismatch: expected USD, got EUR")
	})
}

func TestCost_MulDecimal(t *testing.T) {
	c := cost.NewMonthly(decimal.NewFromFloat(1.23), "USD")
	d := decimal.NewFromInt(3)
	assertDecimalEqual(t, decimal.NewFromFloat(3.69), c.MulDecimal(d).Decimal)
}

func assertDecimalEqual(t *testing.T, expected, actual decimal.Decimal) {
	assert.Truef(t, expected.Equal(actual), "Not equal:\nexpected: %s\nactual  : %s", expected, actual)
}
