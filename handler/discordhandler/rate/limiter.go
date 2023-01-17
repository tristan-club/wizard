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
	//idLimiter = msglimiter.NewIdRateLimiter(rate.Limit(10), 20)
	globalLimiter = rate.NewLimiter(rate.Limit(40), 40)
}

func CheckLimit(chatId string) {
	globalLimiter.Wait(context.Background())
	//if chatId != "" {
	//	cid, err := strconv.ParseInt(chatId, 10, 64)
	//	if err == nil {
	//		idLimiter.GetLimiter(cid).Wait(context.Background())
	//	}
	//}
}
