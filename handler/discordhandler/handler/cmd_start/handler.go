package cmd_start

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/util"
)

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options:       nil,
		Version:       "1",
	},
	Handler: startSendHandler,
}

func startSendHandler(ctx *dcontext.Context) error {

	var user *controller_pb.User
	var isCreateAccount bool
	var pinCode string

	getUserResp, err := ctx.CM.GetUser(ctx.Context, &controller_pb.GetUserReq{
		OpenId:   ctx.Requester.RequesterOpenId,
		OpenType: ctx.Requester.RequesterOpenType,
	})
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "get user",
			"error":  err.Error(),
			"openId": ctx.Requester.RequesterOpenId,
		}).Send()
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if getUserResp.CommonResponse.Code != he.Success {
		if getUserResp.CommonResponse.Code == pconst.CODE_USER_NOT_EXIST {

		} else {
			return tcontext.RespToError(getUserResp.CommonResponse)
		}
	} else {
		user = getUserResp.Data
	}

	if user == nil {
		pinCode = util.GenerateUuid(true)[:6]
		addUserResp, err := ctx.CM.AddUser(ctx.Context, &controller_pb.AddUserReq{
			OpenId:        ctx.Requester.RequesterOpenId,
			OpenType:      int32(ctx.Requester.RequesterOpenType),
			IsOpenInit:    true,
			CreateAccount: true,
			PinCode:       pinCode,
			ChannelId:     ctx.GetGroupChannelId(),
			Username:      ctx.GetUserName(),
			Nickname:      ctx.GetNickname(),
		})
		if err != nil {
			log.Error().Fields(map[string]interface{}{
				"action": "get user",
				"error":  err.Error(),
				"openId": ctx.Requester.RequesterOpenId,
			}).Send()
			return he.NewServerError(he.CodeWalletRequestError, "", err)
		} else if addUserResp.CommonResponse.Code != he.Success {
			return tcontext.RespToError(addUserResp.CommonResponse)
		} else {
			log.Debug().Fields(map[string]interface{}{"action": "init user", "userNo": addUserResp.Data.UserNo, "username": ctx.GetAvailableName(), "address": addUserResp.Data.DefaultAccountAddr, "pinCode": pinCode}).Send()
			log.Info().Fields(map[string]interface{}{"action": "init user", "userNo": addUserResp.Data.UserNo, "username": ctx.GetAvailableName(), "address": addUserResp.Data.DefaultAccountAddr}).Send()
			isCreateAccount = true
			user = addUserResp.Data
		}
	}

	walletContent := "⚡️ Wallet\n"
	if isCreateAccount {
		walletContent += fmt.Sprintf(text.CreateAccountSuccess, user.DefaultAccountAddr, pinCode)
		walletContent = fmt.Sprintf("%s\n%s", walletContent, text.MessageDisappearSoon)
	} else {
		walletContent += fmt.Sprintf(text.GetAccountSuccess, user.DefaultAccountAddr)
	}

	if text.CustomStartMenu != "" {
		walletContent = fmt.Sprintf("%s\n%s", text.CustomStartMenu, walletContent)

	}

	err = ctx.FollowUpReply(walletContent)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send mst", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}

	return nil

}
