package expire_message

import (
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pkg/log"
)

func ClearPreviousStepExpireMessage(ctx *tcontext.Context) {
	expireMessages := expiremessage_state.GetExpireMessage(ctx.OpenId())
	for _, expireMessage := range expireMessages {
		herr := ctx.DeleteMessage(expireMessage.ChatId, expireMessage.MessageId)
		if herr != nil {
			log.Error().Fields(map[string]interface{}{"action": "delete expire message error", "error": herr.Error(), "messageId": expireMessage.MessageId}).Send()
		} else {
			log.Debug().Fields(map[string]interface{}{"action": "delete message", "messageId": expireMessage.MessageId}).Send()
		}
	}
	expiremessage_state.ClearExpireMessage(ctx.OpenId())
}
