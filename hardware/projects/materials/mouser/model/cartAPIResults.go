package model

// CartAPIResults is an object returned after calling the Cart API
type CartAPIResults struct {
	Errors           []APIError `json:"Errors"`
	CartKey          string     `json:"CartKey"`
	CurrencyCode     string     `json:"CurrencyCode"`
	CartItems        []CartItem `json:"CartItems"`
	TotalItemCount   APIUint    `json:"TotalItemCount"`
	MerchandiseTotal APIFloat   `json:"MerchandiseTotal"`
}
