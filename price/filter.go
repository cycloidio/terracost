package price

// Filter is used to filter prices.
type Filter struct {
	Unit             *string
	Currency         *string
	AttributeFilters []*AttributeFilter
}

// AttributeFilter is used for filtering of prices by attribute.
type AttributeFilter struct {
	Key        string
	Value      *string
	ValueRegex *string
}
