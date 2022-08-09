package cmd_my_wallet

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	he "github.com/tristan-club/wizard/pkg/error"
)

var Handler *chain.ChainHandler

const (
	myWalletBaseUri = "https://tristanclub.com/wallet/my?address=%s&app_id=TestAppId&access_token=%s"
	myWalletContent = "💰 *This is your wallet*\n_↓↓↓Click To View  Details↓↓↓_\n"
	buttonContent   = "👉🏻Show Detail"
)

func init() {
	Handler = chain.NewChainHandler(cmd.CmdMyWallet, myWalletHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom)
}

func myWalletHandler(ctx *tcontext.Context) error {
	if !ctx.U.FromChat().IsPrivate() {
		return nil
	}

	//defer userstate.ResetState(ctx.Requester.RequesterOpenId)

	var currUri string
	if reps, err := ctx.CM.InitAccessToken(ctx.Context, &controller_pb.InitAccessTokenReq{UserId: ctx.Requester.RequesterUserNo}); err != nil {
		log.Error().Msgf("init accessToken error:%s", err)
	} else {
		if reps.CommonResponse.Code != he.Success {
			log.Error().Msgf("init accessToken error:%s", reps.CommonResponse.Message)
		} else {
			currUri = fmt.Sprintf(myWalletBaseUri, ctx.Requester.RequesterDefaultAddress, reps.Data.Token)
		}
	}

	if len(currUri) == 0 {
		ctx.Send(ctx.U.FromChat().ID, "init access token failed", nil, false, false)
		return nil
	}

	km := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL(buttonContent, currUri)})

	if thisMsg, herr := ctx.Send(ctx.U.FromChat().ID, myWalletContent, km, true, false); herr != nil {
		log.Error().Msgf("send my wallet inlineKeyBoard error:%s", herr.Msg())
		return herr
	} else {
		ctx.SetDeadlineMsg(ctx.U.FromChat().ID, thisMsg.MessageID, pconst.COMMON_KEYBOARD_DEADLINE)
	}

	return nil
}
