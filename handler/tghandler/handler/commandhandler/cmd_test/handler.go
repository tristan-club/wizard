package cmd_test

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
)

var Handler = chain.NewChainHandler(cmd.CmdTest, TestHandler).AddPreHandler(prehandler.UserMustBeAdmin)

func TestHandler(ctx *tcontext.Context) error {
	log.Info().Fields(map[string]interface{}{"action": "ss"}).Send()

	photo, err := ctx.BotApi.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{UserID: ctx.U.SentFrom().ID})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "err", "error": err.Error()}).Send()
	}
	a := photo.Photos[0][0]

	_, err = ctx.BotApi.Send(tgbotapi.NewChatPhoto(5361549971, tgbotapi.FileID(a.FileID)))
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "error", "error": err.Error()}).Send()
	}
	_, herr := ctx.Send(ctx.U.SentFrom().ID, fmt.Sprintf(text.GetAccountSuccess, ctx.Requester.RequesterDefaultAddress), nil, true, false)
	return herr
}
