package prehandler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
)

func BotMustBeAdmin(ctx *tcontext.Context) error {
	if !ctx.U.FromChat().IsPrivate() {
		message := ctx.U.Message
		if message == nil && ctx.U.CallbackQuery != nil {
			message = ctx.U.CallbackQuery.Message
		}
		if message == nil {
			log.Error().Fields(map[string]interface{}{"action": "invalid bot must admin handler config ", "ctx": ctx}).Send()
			return nil
		}

		chatMember, err := ctx.BotApi.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID:             ctx.U.FromChat().ID,
			SuperGroupUsername: "",
			UserID:             ctx.BotId,
		}})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "get chat member error", "error": err.Error()}).Send()
			return he.NewServerError(he.ServerError, "", err)
		} else if !chatMember.IsAdministrator() && !chatMember.IsCreator() {
			log.Info().Fields(map[string]interface{}{"action": "bot not admin", "chatMember": chatMember}).Send()
			return he.NewBusinessError(pconst.CodePermissionRefused, "Bot needs admin rights to perform this action. ", nil)
		}
	}
	return nil
}
