package mouser

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	// MaxCallsPerMinute is the maximum calls allowed by the Mouser API per minute
	MaxCallsPerMinute = 30
)

// RateLimiter is a struct to limit number of calls to an outside API
type RateLimiter struct {
	sync.Mutex
	tokens     uint32
	rate       time.Duration
	nextUpdate time.Time
}

// NewRateLimiter creates a new RateLimiter object with `toks` initial number of tokens and adding `1` token every `r` duration
func NewRateLimiter(toks uint32, r time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens: toks,
		rate:   r,
	}

	rl.start()
	return rl
}

func (r *RateLimiter) start() {
	go func() {
		for {
			now := time.Now()
			if r.nextUpdate.Before(now) {
				r.nextUpdate = now.Add(r.rate)

				r.Lock()
				tok := atomic.LoadUint32(&r.tokens)
				if tok < MaxCallsPerMinute {
					atomic.AddUint32(&r.tokens, 1)
				}
				r.Unlock()
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// Allowed returns true if access is allowed, false otherwise
func (r *RateLimiter) Allowed() bool {
	r.Lock()
	defer r.Unlock()

	tok := atomic.LoadUint32(&r.tokens)
	if tok > 0 {
		atomic.AddUint32(&r.tokens, ^uint32(0)) // decrement the number of tokens available
		return true
	}

	return false
}

func (r *RateLimiter) getTokens() uint32 {
	r.Lock()
	defer r.Unlock()

	return atomic.LoadUint32(&r.tokens)
}
