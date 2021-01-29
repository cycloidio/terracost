package mysql_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/mysql"
	"github.com/cycloidio/terracost/product"
)

var productColumns = []string{"id", "provider", "sku", "service", "family", "location", "attributes"}

func TestProductRepository_FindByVendorAndSKU(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := mysql.NewProductRepository(db)

	rows := mock.NewRows(productColumns).AddRow(1, "aws", "PRODUCT", "service", "family", "location", `{"key":"value"}`)
	mock.ExpectQuery(`SELECT .+ FROM .+ WHERE provider = \? AND sku = \? LIMIT 1`).
		WithArgs("aws", "PRODUCT").
		WillReturnRows(rows)

	prod, err := repo.FindByVendorAndSKU(context.Background(), "aws", "PRODUCT")
	require.NoError(t, err)

	expected := &product.Product{
		ID:         1,
		Provider:   "aws",
		SKU:        "PRODUCT",
		Service:    "service",
		Family:     "family",
		Location:   "location",
		Attributes: map[string]string{"key": "value"},
	}

	require.Equal(t, expected, prod)
}

func TestProductRepository_Filter(t *testing.T) {
	t.Run("NoFilters", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := mysql.NewProductRepository(db)

		rows := mock.NewRows(productColumns).AddRow(1, "aws", "PRODUCT", "service", "family", "location", `{"key":"value"}`)
		mock.ExpectQuery(`SELECT .+ FROM .+`).WillReturnRows(rows)

		prods, err := repo.Filter(context.Background(), &product.Filter{})
		require.NoError(t, err)

		expected := []*product.Product{
			{
				ID:         1,
				Provider:   "aws",
				SKU:        "PRODUCT",
				Service:    "service",
				Family:     "family",
				Location:   "location",
				Attributes: map[string]string{"key": "value"},
			},
		}

		require.Equal(t, expected, prods)
	})

	t.Run("ColumnFilters", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := mysql.NewProductRepository(db)

		rows := mock.NewRows(productColumns).AddRow(1, "aws", "PRODUCT", "service", "family", "location", `{"key":"value"}`)
		mock.ExpectQuery(`SELECT .+ FROM .+ WHERE provider = \? AND location = \? AND service = \? AND family = \?`).
			WithArgs("aws", "location", "service", "family").
			WillReturnRows(rows)

		filter := &product.Filter{
			Provider: strPtr("aws"),
			Service:  strPtr("service"),
			Family:   strPtr("family"),
			Location: strPtr("location"),
		}
		prods, err := repo.Filter(context.Background(), filter)
		require.NoError(t, err)

		expected := []*product.Product{
			{
				ID:         1,
				Provider:   "aws",
				SKU:        "PRODUCT",
				Service:    "service",
				Family:     "family",
				Location:   "location",
				Attributes: map[string]string{"key": "value"},
			},
		}

		require.Equal(t, expected, prods)
	})

	t.Run("AttributeFilters", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		repo := mysql.NewProductRepository(db)

		rows := mock.NewRows(productColumns).AddRow(1, "aws", "PRODUCT", "service", "family", "location", `{"key":"value","other":"value2"}`)
		mock.ExpectQuery(`SELECT .+ FROM .+ WHERE JSON_UNQUOTE\(JSON_EXTRACT\(attributes, '\$\.key'\)\) = \? AND JSON_UNQUOTE\(JSON_EXTRACT\(attributes, '\$\.other'\)\) RLIKE \?`).
			WithArgs("value", "lue").
			WillReturnRows(rows)

		filter := &product.Filter{
			AttributeFilters: []*product.AttributeFilter{
				{Key: "key", Value: strPtr("value")},
				{Key: "other", ValueRegex: strPtr("lue")},
			},
		}
		prods, err := repo.Filter(context.Background(), filter)
		require.NoError(t, err)

		expected := []*product.Product{
			{
				ID:       1,
				Provider: "aws",
				SKU:      "PRODUCT",
				Service:  "service",
				Family:   "family",
				Location: "location",
				Attributes: map[string]string{
					"key":   "value",
					"other": "value2",
				},
			},
		}

		require.Equal(t, expected, prods)
	})
}

func TestProductRepository_Upsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := mysql.NewProductRepository(db)

	mock.ExpectExec(`INSERT INTO .+ VALUES .+ ON DUPLICATE KEY UPDATE .+`).
		WithArgs("aws", "PRODUCT", "service", "family", "location", `{"key":"value"}`).
		WillReturnResult(sqlmock.NewResult(123, 1))

	prod := &product.Product{
		Provider:   "aws",
		SKU:        "PRODUCT",
		Service:    "service",
		Family:     "family",
		Location:   "location",
		Attributes: map[string]string{"key": "value"},
	}

	id, err := repo.Upsert(context.Background(), prod)
	require.NoError(t, err)
	require.Equal(t, product.ID(123), id)
}

func strPtr(s string) *string {
	return &s
}
