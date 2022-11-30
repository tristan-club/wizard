package msglimiter

import (
	"golang.org/x/time/rate"
	"sync"
)

type Limiter struct {
	keys map[int64]*rate.Limiter
	mu   *sync.RWMutex
	r    rate.Limit
	b    int
}

func NewIdRateLimiter(r rate.Limit, b int) *Limiter {
	i := &Limiter{
		keys: make(map[int64]*rate.Limiter),
		mu:   &sync.RWMutex{},
		r:    r,
		b:    b,
	}

	return i
}

// Add creates a new rate limiter and adds it to the keys map,
// using the key
func (i *Limiter) Add(key int64) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)

	i.keys[key] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided key if it exists.
// Otherwise, calls Add to add key address to the map
func (i *Limiter) GetLimiter(key int64) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.keys[key]

	if !exists {
		i.mu.Unlock()
		return i.Add(key)
	}

	i.mu.Unlock()

	return limiter
}
