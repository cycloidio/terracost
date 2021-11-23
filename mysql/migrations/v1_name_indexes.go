package migrations

// v1NameIndexes will rename all the indexes from the default value
// to another specific name which we can check.
// It'll also add the location to the pricing_products UQ index
var v1NameIndexes = Migration{
	Name: "Name Indexes",
	SQL: `
		ALTER TABLE pricing_products
			DROP INDEX provider;
		ALTER TABLE pricing_products
			ADD CONSTRAINT UNIQUE uq__provider__sku__location (provider, sku, location);

		ALTER TABLE pricing_products
			DROP INDEX provider_2;
		ALTER TABLE pricing_products
			ADD INDEX idx__provider__location__service__family (provider, location, service, family);

		ALTER TABLE pricing_product_prices
			DROP FOREIGN KEY pricing_product_prices_ibfk_1;
		ALTER TABLE pricing_product_prices
			DROP INDEX product_id;
		ALTER TABLE pricing_product_prices
			ADD CONSTRAINT fk__pricing_product_prices__pricing_products FOREIGN KEY (product_id) REFERENCES pricing_products (id);
		ALTER TABLE pricing_product_prices
			ADD CONSTRAINT UNIQUE uq__product_id__hash (product_id, hash);
	`,
}
