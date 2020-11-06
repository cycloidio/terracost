package product

// ID represents the Product ID.
type ID uint32

// Product is an entry of a single SKU.
type Product struct {
	ID         ID
	Provider   string
	SKU        string
	Service    string
	Family     string
	Location   string
	Attributes map[string]string
}
