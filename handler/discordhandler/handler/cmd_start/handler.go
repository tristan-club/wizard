package cmd_start

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/util"
	"strings"
)

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options:       nil,
		Version:       "1",
	},
	Handler: startSendHandler,
}

type StartParam struct {
	IgnoreGetAccountMsg bool `json:"ignore_get_account_msg"`
	IgnoreGuideMsg      bool `json:"ignore_guide_msg"`
}

type StartResult struct {
	UserId         string             `json:"user_id"`
	Address        string             `json:"address"`
	TemporaryToken string             `json:"temporary_token"`
	CreateAddress  bool               `json:"create_address"`
	StartContent   string             `json:"start_content"`
	StartMsg       *discordgo.Message `json:"start_msg"`
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
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if getUserResp.CommonResponse.Code != he.Success {
		if getUserResp.CommonResponse.Code == pconst.CODE_USER_NOT_EXIST {

		} else {
			log.Error().Fields(map[string]interface{}{"action": "get user error", "error": getUserResp}).Send()
			return tcontext.RespToError(getUserResp.CommonResponse)
		}
	} else {
		user = getUserResp.Data
	}
	param := &StartParam{}
	if !util.IsNil(ctx.Param) {
		if _param, ok := ctx.Param.(*StartParam); ok {
			log.Info().Fields(map[string]interface{}{"action": "get start param", "param": ctx.Param}).Send()
			param = _param
		}
	}
	result := &StartResult{}

	if user == nil {
		//pinCode = util.GenerateUuid(true)[:6]
		pinCode = pconst.DefaultPinCode
		addUserResp, err := ctx.CM.AddUser(ctx.Context, &controller_pb.AddUserReq{
			OpenId:        ctx.Requester.RequesterOpenId,
			OpenType:      int32(ctx.Requester.RequesterOpenType),
			IsOpenInit:    true,
			CreateAccount: true,
			PinCode:       pinCode,
			ChannelId:     ctx.GetGroupChannelId(),
			Username:      ctx.GetUserName(),
			Nickname:      ctx.GetNickname(),
			AppId:         ctx.Requester.RequesterAppId,
		})
		if err != nil {
			log.Error().Fields(map[string]interface{}{
				"action": "get user",
				"error":  err.Error(),
				"openId": ctx.Requester.RequesterOpenId,
			}).Send()
			return he.NewServerError(pconst.CodeWalletRequestError, "", err)
		} else if addUserResp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "add user error", "error": addUserResp}).Send()
			return tcontext.RespToError(addUserResp.CommonResponse)
		} else {
			log.Debug().Fields(map[string]interface{}{"action": "init user", "userNo": addUserResp.Data.UserNo, "username": ctx.GetAvailableName(), "address": addUserResp.Data.DefaultAccountAddr, "pinCode": pinCode}).Send()
			log.Info().Fields(map[string]interface{}{"action": "init user", "userNo": addUserResp.Data.UserNo, "username": ctx.GetAvailableName(), "address": addUserResp.Data.DefaultAccountAddr}).Send()
			isCreateAccount = true
			user = addUserResp.Data
		}
	}

	result.UserId = user.UserNo
	result.Address = user.DefaultAccountAddr
	result.CreateAddress = isCreateAccount

	var components []discordgo.MessageComponent

	if config.UseTemporaryToken() {
		var temporaryToken string
		initTemporaryTokenResp, err := ctx.CM.InitTemporaryToken(ctx.Context, &controller_pb.InitTemporaryTokenReq{
			UserId: user.UserNo,
			AppId:  ctx.Requester.RequesterAppId,
		})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "request controller svc error", "error": err.Error()}).Send()
			//return he.NewServerError(pconst.CodeWalletRequestError, "", err)
		} else if initTemporaryTokenResp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "init temporary token error", "error": initTemporaryTokenResp}).Send()
			//return he.NewServerError(int(initTemporaryTokenResp.CommonResponse.Code), "", fmt.Errorf(initTemporaryTokenResp.CommonResponse.Message))
		} else {
			temporaryToken = initTemporaryTokenResp.Data.Token
		}

		suffix := fmt.Sprintf("?app_id=%s&bot_type=%d&token=%s", ctx.Requester.RequesterAppId, pconst.PlatformDiscord, temporaryToken)
		suffix = strings.ReplaceAll(suffix, " ", "%20")

		result.TemporaryToken = temporaryToken

		components = []discordgo.MessageComponent{
			&discordgo.Button{
				Label:    text.KBAccount,
				Style:    discordgo.LinkButton,
				Disabled: false,
				Emoji:    discordgo.ComponentEmoji{},
				URL:      fmt.Sprintf("%s%s", pconst.WebAppMenuUrl, suffix),
			},
			&discordgo.Button{
				Label:    text.KBActivity,
				Style:    discordgo.LinkButton,
				Disabled: false,
				Emoji:    discordgo.ComponentEmoji{},
				URL:      fmt.Sprintf("%s%s", pconst.WebAppActivityUrl, suffix),
			},
			&discordgo.Button{
				Label:    text.KBCAT,
				Style:    discordgo.LinkButton,
				Disabled: false,
				Emoji:    discordgo.ComponentEmoji{},
				URL:      fmt.Sprintf("%s%s", pconst.WebAppCAT, suffix),
			},
		}
	}

	var dmContent string
	respContent := "⚡️ Wallet\n"
	if isCreateAccount {
		respContent += fmt.Sprintf(text.CreateAccountSuccess, user.DefaultAccountAddr, pinCode)

		if !ctx.IsPrivate() {
			dmContent = respContent
			respContent = fmt.Sprintf("%s\n%s", respContent, text.MessageDisappearSoon)
		}

	} else {
		respContent += fmt.Sprintf(text.GetAccountSuccess, user.DefaultAccountAddr)
	}

	//if text.CustomStartMenu != "" {
	//	respContent = fmt.Sprintf("%s\n%s", text.CustomStartMenu, respContent)
	//}

	result.StartContent = respContent

	if !param.IgnoreGuideMsg {

		wp := &discordgo.WebhookParams{Embeds: []*discordgo.MessageEmbed{{Description: respContent}}}
		if len(components) > 0 {
			wp.Components = []discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
		}

		result.StartMsg, err = ctx.FollowUpReplyComplex(wp)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
			return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
		}
	}

	if dmContent != "" {
		err = ctx.DM(dmContent)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "bot send msg error", "error": err.Error()}).Send()
		}
	}

	ctx.Result = result

	return nil

}
