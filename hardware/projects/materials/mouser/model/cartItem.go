package model

// CartItem represents a single cart item as seen by the Mouser API
type CartItem struct {
	Errors            []APIError `json:"Errors"`
	MouserATS         APIUint    `json:"MouserATS"`
	Quantity          APIUint    `json:"Quantity"`
	PartsPerReel      APIUint    `json:"PartsPerReel"`
	ScheduledReleases []struct {
		Key   string  `json:"key"`
		Value APIUint `json:"value"`
	} `json:"ScheduledReleases"`
	InfoMessage            []string `json:"InfoMessage"`
	MouserPartNumber       string   `json:"MouserPartNumber"`
	MfrPartNumber          string   `json:"MfrPartNumber"`
	Description            string   `json:"Description"`
	CartItemCustPartNumber string   `json:"CartItemCustPartNumber"`
	UnitPrice              APIFloat `json:"UnitPrice"`
	ExtendedPrice          APIFloat `json:"ExtendedPrice"`
	LifeCycle              string   `json:"LifeCycle"`
	Manufacturer           string   `json:"Manufacturer"`
	SalesMultipleQty       APIUint  `json:"SalesMultipleQty"`
	SalesMinimumOrderQty   APIUint  `json:"SalesMinimumOrderQty"`
}
