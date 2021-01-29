package mysql_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/price"
	"github.com/cycloidio/terracost/product"
)

var priceColumns = []string{"id", "hash", "product_id", "currency", "price", "unit", "attributes"}

func TestPriceRepository_FilterByProduct(t *testing.T) {
	t.Run("NoFilters", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := mysql.NewPriceRepository(db)

		rows := mock.NewRows(priceColumns).AddRow(1, "HASH", 1, "USD", decimal.RequireFromString("1.23"), "Hrs", `{"key":"value"}`)
		mock.ExpectQuery(`SELECT .+ FROM .+ WHERE product_id = \?`).
			WithArgs(1).
			WillReturnRows(rows)

		prices, err := repo.Filter(context.Background(), product.ID(1), nil)
		require.NoError(t, err)

		expected := []*price.Price{
			{
				ID:         1,
				Unit:       "Hrs",
				Currency:   "USD",
				Value:      decimal.RequireFromString("1.23"),
				Attributes: map[string]string{"key": "value"},
			},
		}

		require.Equal(t, expected, prices)
	})

	t.Run("ColumnFilters", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := mysql.NewPriceRepository(db)

		rows := mock.NewRows(priceColumns).AddRow(1, "HASH", 1, "USD", decimal.RequireFromString("1.23"), "Hrs", `{"key":"value"}`)
		mock.ExpectQuery(`SELECT .+ FROM .+ WHERE product_id = \? AND unit = \? AND currency = \?`).
			WithArgs(1, "Hrs", "USD").
			WillReturnRows(rows)

		filter := &price.Filter{
			Unit:     strPtr("Hrs"),
			Currency: strPtr("USD"),
		}
		prices, err := repo.Filter(context.Background(), product.ID(1), filter)
		require.NoError(t, err)

		expected := []*price.Price{
			{
				ID:         1,
				Unit:       "Hrs",
				Currency:   "USD",
				Value:      decimal.RequireFromString("1.23"),
				Attributes: map[string]string{"key": "value"},
			},
		}

		require.Equal(t, expected, prices)
	})

	t.Run("AttributeFilters", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := mysql.NewPriceRepository(db)

		rows := mock.NewRows(priceColumns).AddRow(1, "HASH", 1, "USD", decimal.RequireFromString("1.23"), "Hrs", `{"key":"value","other":"value2"}`)
		mock.ExpectQuery(`SELECT .+ FROM .+ WHERE product_id = \? AND JSON_UNQUOTE\(JSON_EXTRACT\(attributes, '\$\.key'\)\) = \? AND JSON_UNQUOTE\(JSON_EXTRACT\(attributes, '\$\.other'\)\) RLIKE \?`).
			WithArgs(1, "value", "lue").
			WillReturnRows(rows)

		filter := &price.Filter{
			AttributeFilters: []*price.AttributeFilter{
				{Key: "key", Value: strPtr("value")},
				{Key: "other", ValueRegex: strPtr("lue")},
			},
		}
		prices, err := repo.Filter(context.Background(), product.ID(1), filter)
		require.NoError(t, err)

		expected := []*price.Price{
			{
				ID:       1,
				Unit:     "Hrs",
				Currency: "USD",
				Value:    decimal.RequireFromString("1.23"),
				Attributes: map[string]string{
					"key":   "value",
					"other": "value2",
				},
			},
		}

		require.Equal(t, expected, prices)
	})
}
