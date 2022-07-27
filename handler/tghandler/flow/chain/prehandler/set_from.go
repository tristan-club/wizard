package prehandler

import (
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
)

func SetFrom(ctx *tcontext.Context) error {
	fromParams := map[string]interface{}{
		"user_no": ctx.Requester.RequesterUserNo,
		"from":    ctx.Requester.RequesterDefaultAddress,
	}

	if !ctx.U.FromChat().IsPrivate() {
		fromParams["channel_id"] = ctx.Requester.RequesterChannelId
	}

	userstate.BatchSaveParam(ctx.OpenId(), fromParams)

	return nil
}
