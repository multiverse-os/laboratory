package mouser

// ErrorRateLimited is returned when API is answering a TooManyRequests error
type ErrorRateLimited struct {
	error
}
