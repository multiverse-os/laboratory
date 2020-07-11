package mouser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bom-maker/mouser/model"
)

const (
	searchAPIEndpoint     = "search"
	searchAPIVersion      = 1
	rateLimitedRetryEvery = 2 * time.Second
)

type searchQuery struct {
	SearchByPartRequest struct {
		MouserPartNumber string `json:"mouserPartNumber"`
	} `json:"SearchByPartRequest"`
}

func newSearchQuery(mouserRef string) searchQuery {
	var query searchQuery
	query.SearchByPartRequest.MouserPartNumber = mouserRef
	return query
}

// SearchByPartNumber searches by part reference
func (a *API) SearchByPartNumber(mouserRef string) (*model.Part, error) {
	var err error

	query := newSearchQuery(mouserRef)
	reqBody, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	url := a.buildAPIUrl(fmt.Sprintf("%s/partnumber", searchAPIEndpoint), searchAPIVersion)

	// Wait for a token
	a.waitForBeingAllowed()
	//fmt.Printf("\tAllowed to call API, tokens=%d\n", a.RateLimiter.getTokens())

	resp, err := http.Post(url, ContentType, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//fmt.Println(string(body))

	results := model.SearchAPIResults{}
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, err
	}

	if len(results.Errors) > 0 {
		// Checking for rate limited error
		for _, e := range results.Errors {
			if e.Code == "TooManyRequests" {
				return nil, ErrorRateLimited{err}
			}
		}

		return nil, fmt.Errorf("An error occurred calling API, err=%v", results.Errors)
	}

	// Looking for reference into results' parts
	for _, part := range results.SearchResults.Parts {
		// We are looking for an exact match !
		if part.MouserPartNumber == mouserRef {
			return &part, nil
		}
	}

	return nil, fmt.Errorf("Cannot find any part, mouserRef=%s", mouserRef)
}
