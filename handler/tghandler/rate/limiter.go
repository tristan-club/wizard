package rate

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

type Limiter struct {
	keys map[int64]*rate.Limiter
	mu   *sync.RWMutex
	r    rate.Limit
	b    int
}

var idLimiter *Limiter
var globalLimiter *rate.Limiter

// NewLimiter creates both chat and global rate limiters.
func init() {
	idLimiter = newIdRateLimiter(rate.Limit(1), 19)
	globalLimiter = rate.NewLimiter(rate.Limit(30), 30)
}

// NewRateLimiter .
func newIdRateLimiter(r rate.Limit, b int) *Limiter {
	i := &Limiter{
		keys: make(map[int64]*rate.Limiter),
		mu:   &sync.RWMutex{},
		r:    r,
		b:    b,
	}

	return i
}

func CheckLimit(chatId int64) {
	globalLimiter.Wait(context.Background())
	idLimiter.GetLimiter(chatId).Wait(context.Background())
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
