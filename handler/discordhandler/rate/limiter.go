package rate

import (
	"context"
	"github.com/tristan-club/wizard/handler/msglimiter"
	"golang.org/x/time/rate"
	"strconv"
)

var idLimiter *msglimiter.Limiter
var globalLimiter *rate.Limiter

// NewLimiter creates both chat and global rate limiters.
func init() {
	idLimiter = msglimiter.NewIdRateLimiter(rate.Limit(1), 20)
	globalLimiter = rate.NewLimiter(rate.Limit(50), 50)
}

func CheckLimit(chatId string) {
	globalLimiter.Wait(context.Background())
	if chatId != "" {
		cid, err := strconv.ParseInt(chatId, 10, 64)
		if err == nil {
			idLimiter.GetLimiter(cid).Wait(context.Background())
		}
	}
}
