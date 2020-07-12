package mouser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bom-maker/core"
	"github.com/bom-maker/mouser/model"
)

const (
	cartAPIEndpoint = "cart"
	cartAPIVersion  = 1
)

type cartQueryItem struct {
	MouserPartNumber   string        `json:"MouserPartNumber"`
	Quantity           model.APIUint `json:"Quantity"`
	CustomerPartNumber string        `json:"CustomerPartNumber,omitempty"`
}

type cartQuery struct {
	CartKey   string          `json:"CartKey,omitempty"`
	CartItems []cartQueryItem `json:"CartItems"`
}

// newCartQuery creates a new query to handle carts
func newCartQuery(parts []core.UberPart, multiplier int) cartQuery {
	cq := cartQuery{}

	for _, part := range parts {
		cq.CartItems = append(cq.CartItems, cartQueryItem{
			MouserPartNumber: part.MouserPartNumber,
			Quantity:         part.Quantity * model.APIUint(multiplier),
		})
	}

	return cq
}

// InsertItemsInCart inserts items on an existing cart or create a new one if cartKey is not defined
func (a *API) InsertItemsInCart(cartKey string, parts []core.UberPart, multiplier int) (*model.CartAPIResults, error) {
	var err error

	query := newCartQuery(parts, multiplier)
	reqBody, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	url := a.buildAPIUrl(fmt.Sprintf("%s/items/insert", cartAPIEndpoint), cartAPIVersion)

	resp, err := http.Post(url, ContentType, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	results := model.CartAPIResults{}
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, err
	}

	if len(results.Errors) > 0 {
		return &results, fmt.Errorf("An error occurred during API call, check for error in the returning object")
	}

	return &results, nil
}
