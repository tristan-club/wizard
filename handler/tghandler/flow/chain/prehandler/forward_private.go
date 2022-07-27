package prehandler

import (
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/inline_keybord"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pkg/log"
)

func ForwardPrivate(ctx *tcontext.Context) error {
	if !ctx.U.FromChat().IsPrivate() {
		var EnvelopeTypeKeyboard, deadlineTime = inline_keybord.NewForwardPrivateKeyBoard(ctx)
		if replyMsg, herr := ctx.Reply(ctx.U.FromChat().ID, text.SwitchPrivate, EnvelopeTypeKeyboard, false); herr == nil {
			inline_keybord.DeleteDeadKeyboard(ctx, deadlineTime, &replyMsg)
		} else {
			log.Error().Msgf("forward private error:%s", herr.Error())
		}
	}
	return nil
}
