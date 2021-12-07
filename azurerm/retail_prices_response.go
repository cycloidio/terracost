package azurerm

type retailPricesResponse struct {
	BillingCurrency    string        `json:"BillingCurrency"`
	CustomerEntityID   string        `json:"CustomerEntityId"`
	CustomerEntityType string        `json:"CustomerEntityType"`
	Items              []retailPrice `json:"Items"`
	NextPageLink       string        `json:"NextPageLink"`
	Count              int           `json:"Count"`
}

type retailPrice struct {
	CurrencyCode         string  `json:"currencyCode"`
	TierMinimumUnits     float64 `json:"tierMinimumUnits"`
	RetailPrice          float64 `json:"retailPrice"`
	UnitPrice            float64 `json:"unitPrice"`
	ArmRegionName        string  `json:"armRegionName"`
	Location             string  `json:"location"`
	EffectiveStartDate   string  `json:"effectiveStartDate"`
	MeterID              string  `json:"meterId"`
	MeterName            string  `json:"meterName"`
	ProductID            string  `json:"productId"`
	SkuID                string  `json:"skuId"`
	ProductName          string  `json:"productName"`
	SkuName              string  `json:"skuName"`
	ServiceName          string  `json:"serviceName"`
	ServiceID            string  `json:"serviceId"`
	ServiceFamily        string  `json:"serviceFamily"`
	UnitOfMeasure        string  `json:"unitOfMeasure"`
	Type                 string  `json:"type"`
	IsPrimaryMeterRegion bool    `json:"isPrimaryMeterRegion"`
	ArmSkuName           string  `json:"armSkuName"`
	EffectiveEndDate     string  `json:"effectiveEndDate,omitempty"`
	ReservationTerm      string  `json:"reservationTerm,omitempty"`
}
