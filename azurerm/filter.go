package azurerm

import (
	"strings"

	"github.com/cycloidio/terracost/price"
)

// IngestionFilter allows control over what pricing data is ingested. Given a price.WithProduct the function returns
// true if the record should be ingested, false if it should be skipped.
type IngestionFilter func(pp *price.WithProduct) bool

// DefaultFilter ingests all the records without filtering.
func DefaultFilter(_ *price.WithProduct) bool {
	return true
}

// MinimalFilter only ingests the supported records, skipping those that would never be used.
func MinimalFilter(pp *price.WithProduct) bool {

	// Ignore Spot and Reserved Virtual Machines
	if pp.Product.Service == "Virtual Machines" && pp.Product.Family == "Compute" {
		if strings.HasSuffix(pp.Product.Attributes["meter_name"], " Spot") || strings.HasSuffix(pp.Product.Attributes["meter_name"], " Low Priority") {
			return false
		}
	}

	return pp.Price.Attributes["type"] == "Consumption"
}
