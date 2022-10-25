package prehandler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
)

func UserMustBeAdmin(ctx *tcontext.Context) error {
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
			UserID:             ctx.U.SentFrom().ID,
		}})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "get chat member error", "error": err.Error()}).Send()
			return he.NewServerError(he.ServerError, "", err)
		} else if !chatMember.IsAdministrator() && !chatMember.IsCreator() {

			appOwnerResp, err := ctx.CM.GetAppOwner(ctx.Context, &controller_pb.GetAppOwnerReq{UserId: ctx.Requester.RequesterUserNo, AppId: ctx.Requester.RequesterAppId})
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "request controller svc error", "error": err.Error(), "ctx": ctx}).Send()
				return he.NewBusinessError(pconst.CodePermissionRefused, "User needs admin rights to perform this action. ", nil)
			} else if appOwnerResp.CommonResponse.Code != he.Success {
				log.Error().Fields(map[string]interface{}{"action": "get app owner error", "error": appOwnerResp}).Send()
				return he.NewBusinessError(pconst.CodePermissionRefused, "User needs admin rights to perform this action. ", nil)
			} else if len(appOwnerResp.Data) != 1 || appOwnerResp.Data[0].UserId != ctx.Requester.RequesterUserNo {
				log.Info().Fields(map[string]interface{}{"action": "user not admin", "chatMember": chatMember}).Send()
				return he.NewBusinessError(pconst.CodePermissionRefused, "User needs admin rights to perform this action. ", nil)
			}
		}
	}
	return nil
}
