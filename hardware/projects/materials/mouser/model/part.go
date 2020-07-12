package model

import (
	"regexp"
	"sort"
	"strings"
)

var (
	// AvailabilitySentences allow to know if a part is in stock or backordered
	// Yes of course, API is responding with localized sentences...
	AvailabilitySentences = map[string]bool{
		"On Order":     false,
		"In Stock":     true,
		"Sur commande": false,
		"En stock":     true,
	}
)

// PriceBreak represents a price depending on quantity ordered
type PriceBreak struct {
	Quantity APIUint `json:"Quantity"`
	Price    string  `json:"Price"`
	Currency string  `json:"Currency"`
}

// Part represents an electronic component from Mouser perspective
type Part struct {
	Availability           string  `json:"Availability"`
	DatasheetURL           string  `json:"DataSheetUrl"`
	Description            string  `json:"Description"`
	FactoryStock           APIUint `json:"FactoryStock"`
	ImagePath              string  `json:"ImagePath"`
	Category               string  `json:"Category"`
	LeadTime               string  `json:"LeadTime"`
	LifecycleStatus        string  `json:"LifecycleStatus"`
	Manufacturer           string  `json:"Manufacturer"`
	ManufacturerPartNumber string  `json:"ManufacturerPartNumber"`
	Min                    APIUint `json:"Min"`
	Mult                   APIUint `json:"Mult"`
	MouserPartNumber       string  `json:"MouserPartNumber"`
	ProductAttributes      []struct {
		AttributeName  string `json:"AttributeName"`
		AttributeValue string `json:"AttributeValue"`
	} `json:"ProductAttributes"`
	PriceBreaks         []PriceBreak `json:"PriceBreaks,omitempty"`
	AlternatePackagings []struct {
		APMfrPN string `json:"APMfrPN"`
	} `json:"AlternatePackagings"`
	ProductDetailURL     string  `json:"ProductDetailUrl"`
	Reeling              bool    `json:"Reeling"`
	ROHSStatus           string  `json:"ROHSStatus"`
	SuggestedReplacement string  `json:"SuggestedReplacement"`
	MultiSimBlue         APIUint `json:"MultiSimBlue"`
	UnitWeightKg         struct {
		UnitWeight APIFloat `json:"UnitWeight"`
	} `json:"UnitWeightKg"`
	StandardCost struct {
		StandardCost APIFloat `json:"Standardcost"`
	} `json:"StandardCost"`
	IsDiscontinued        string `json:"IsDiscontinued"`
	RTM                   string `json:"RTM"`
	MouserProductCategory string `json:"MouserProductCategory"`
	IPCCode               string `json:"IPCCode"`
	SField                string `json:"SField"`
	VNum                  string `json:"VNum"`
	ActualMfrName         string `json:"ActualMfrName"`
	AvailableOnOrder      string `json:"AvailableOnOrder"`
	RestrictionMessage    string `json:"RestrictionMessage"`
	PID                   string `json:"PID"`

	ProductCompliance []struct {
		ComplianceName  string `json:"ComplianceName"`
		ComplianceValue string `json:"ComplianceValue"`
	} `json:"ProductCompliance"`
}

// GetUnitPrice returns the price depending on quantity based on PriceBreaks field
func (p *Part) GetUnitPrice(quantity APIUint) PriceBreak {
	sort.Slice(p.PriceBreaks, func(i, j int) bool {
		return p.PriceBreaks[i].Quantity < p.PriceBreaks[j].Quantity
	})

	var previous PriceBreak
	for _, v := range p.PriceBreaks {
		if v.Quantity > quantity {
			if previous.Quantity != 0 {
				return previous
			}

			return v
		}
		previous = v
	}

	if len(p.PriceBreaks) == 0 {
		return PriceBreak{}
	}

	// Nothing else found, returning first one as a default 🤷‍♂️
	return p.PriceBreaks[0]

}

// GetAvailabilityAsNumber returns number of parts available typed as a number
// API returns a string such as "[0-9]+ In stock"
func (p *Part) GetAvailabilityAsNumber() APIUint {
	return GetAPIUintFromString(p.Availability)
}

// GetUnitPriceAsNumber returns unit price typed as a number
// API returns a string such as "[0-9]+ €"
func (p *Part) GetUnitPriceAsNumber(quantity APIUint) APIFloat {
	return GetAPIFloatFromString(p.GetUnitPrice(quantity).Price)
}

// InStock returns true if a part is in stock or false if backordered.
func (p *Part) InStock() bool {
	re := regexp.MustCompile(`[^0-9]+`)
	sentence := strings.Trim(re.FindString(p.Availability), " ")
	return AvailabilitySentences[sentence]
}
