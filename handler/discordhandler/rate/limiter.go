package rate

import (
	"context"
	"github.com/tristan-club/wizard/handler/msglimiter"
	"golang.org/x/time/rate"
)

var idLimiter *msglimiter.Limiter
var globalLimiter *rate.Limiter

// NewLimiter creates both chat and global rate limiters.
func init() {
	idLimiter = msglimiter.NewIdRateLimiter(rate.Limit(1), 19)
	globalLimiter = rate.NewLimiter(rate.Limit(50), 50)
}

func CheckLimit(chatId int64) {
	globalLimiter.Wait(context.Background())
	idLimiter.GetLimiter(chatId).Wait(context.Background())
}
