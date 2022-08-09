package cmd_my_wallet

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	he "github.com/tristan-club/wizard/pkg/error"
)

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Version:       "1",
	},
	Handler: myWalletHandler,
}

const (
	myWalletBaseUri = "https://tristanclub.com/wallet/my?address=%s&app_id=TestAppId&access_token=%s"
	myWalletContent = "💰 *This is your wallet*\n_↓↓↓Click To View  Details↓↓↓_\n"
	buttonContent   = "👉🏻Show Detail"
)

func myWalletHandler(ctx *dcontext.Context) error {

	var currUri string
	if resp, err := ctx.CM.InitAccessToken(ctx.Context, &controller_pb.InitAccessTokenReq{UserId: ctx.Requester.RequesterUserNo}); err != nil {
		log.Error().Msgf("init accessToken error:%s", err)
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else {
		if resp.CommonResponse.Code != he.Success {
			log.Error().Msgf("init accessToken error:%s", resp.CommonResponse.Message)
			return tcontext.RespToError(resp.CommonResponse)
		} else {
			currUri = fmt.Sprintf(myWalletBaseUri, ctx.Requester.RequesterDefaultAddress, resp.Data.Token)
		}
	}

	if err := ctx.FollowUpReply(fmt.Sprintf("%s\n%s", myWalletContent, currUri)); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}

	return nil
}
