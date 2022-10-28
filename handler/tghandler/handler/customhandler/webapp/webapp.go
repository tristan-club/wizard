package webapp

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
)

var WebAppHandler *chain.ChainHandler

func init() {
	WebAppHandler = chain.NewChainHandler("webapp", webAppSendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom)
}

// todo @mazhonghao webapp
func webAppSendHandler(ctx *tcontext.Context) error {

	initTemporaryTokenResp, err := ctx.CM.InitTemporaryToken(ctx.Context, &controller_pb.InitTemporaryTokenReq{
		UserId: ctx.Requester.RequesterUserNo,
		AppId:  ctx.Requester.RequesterAppId,
	})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "request controller svc error", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if initTemporaryTokenResp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "init temporary token error", "error": initTemporaryTokenResp}).Send()
		return he.NewServerError(int(initTemporaryTokenResp.CommonResponse.Code), "", fmt.Errorf(initTemporaryTokenResp.CommonResponse.Message))
	}

	url := fmt.Sprintf("%s?temporary_token=%s", pconst.WebAppMenuUrl, initTemporaryTokenResp.Data.Token)
	ikm := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{{{Text: pconst.WebAppMenuBtName, WebApp: &tgbotapi.WebAppInfo{
			URL: url,
		}}}},
	}
	_, _ = ctx.Send(ctx.U.SentFrom().ID, text.OpenWebApp, &ikm, false, true)
	_ = ctx.SetChatMenuButton(&tgbotapi.MenuButton{
		Type:   "web_app",
		Text:   "webApp",
		WebApp: &tgbotapi.WebAppInfo{URL: url},
	})

	return nil
}
