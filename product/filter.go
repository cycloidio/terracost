package product

// Filter is used to filter products.
type Filter struct {
	Provider         *string
	SKU              *string
	Service          *string
	Family           *string
	Location         *string
	AttributeFilters []*AttributeFilter
}

// AttributeFilter is used for filtering of products by attribute.
type AttributeFilter struct {
	Key        string
	Value      *string
	ValueRegex *string
}
