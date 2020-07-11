package mouser

import (
	"fmt"
	"time"
)

// https://api.mouser.com/api/docs/ui/index

const (
	// ContentType is the Content-Type used for API calls
	ContentType = "application/json"
	// MouserAPIUrl is the Mouser API base URL
	MouserAPIUrl = "https://api.mouser.com"
)

// API is a struct to represent an object capable of talking to the Mouser API
type API struct {
	APIKey      string
	RateLimiter *RateLimiter
}

// NewAPI returns a new mouser API object
func NewAPI(apiKey string) *API {
	return &API{
		APIKey:      apiKey,
		RateLimiter: NewRateLimiter(MaxCallsPerMinute, 2*time.Second), // 30 tokens max / minute
	}
}

func (a *API) buildAPIUrl(endpoint string, version uint) string {
	return fmt.Sprintf(
		"%s/api/v%d/%s?apiKey=%s",
		MouserAPIUrl,
		version,
		endpoint,
		a.APIKey,
	)
}

func (a API) waitForBeingAllowed() {
	for !a.RateLimiter.Allowed() {
		time.Sleep(100 * time.Millisecond)
	}
}
