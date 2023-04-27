package migrations

// v2ExtendPriceUnit will extend the price unit length as we found
// that some unit in some services are longer that the previous 32
var v2ExtendPriceUnit = Migration{
	Name: "Extend Price Unit field",
	SQL: `
		ALTER TABLE pricing_product_prices
			MODIFY COLUMN unit VARCHAR(255);
	`,
}
