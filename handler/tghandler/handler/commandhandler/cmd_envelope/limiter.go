package cmd_envelope

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/tristan-club/wizard/pkg/util"
	"time"
)

var cacheMgr = cache.New(time.Second*3, time.Second*5)

func checkEnvelopeClaim(envelopeNo, userId string) bool {
	if _, ok := cacheMgr.Get(string(util.HashStr(fmt.Sprintf("%s_%s", envelopeNo, userId)))); ok {
		return true
	}
	cacheMgr.Set(string(util.HashStr(fmt.Sprintf("%s_%s", envelopeNo, userId))), 1, cache.DefaultExpiration)
	return false
}
