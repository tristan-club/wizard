package cmd_start

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/inline_keybord"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/util"
	"time"
)

var Handler = chain.NewChainHandler(cmd.CmdStart, startSendHandler)

var UserChannel = map[string]string{}

func startSendHandler(ctx *tcontext.Context) error {

	if ctx.U.Message == nil {
		log.Error().Fields(map[string]interface{}{"full update context": ctx.U, "warn": "nil message or something"}).Send()
		return he.NewServerError(pconst.CodeInvalidUserState, "", fmt.Errorf("invalid state for start command"))
	}

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
			if ctx.U.FromChat().IsPrivate() {
				channelId := ctx.Requester.GetRequesterChannelId()
				if channelId == "" {
					channelId = UserChannel[ctx.OpenId()]
				}

				pinCode = util.GenerateUuid(true)[:6]
				addUserResp, err := ctx.CM.AddUser(ctx.Context, &controller_pb.AddUserReq{
					OpenId:        ctx.Requester.RequesterOpenId,
					OpenType:      int32(ctx.Requester.RequesterOpenType),
					IsOpenInit:    true,
					CreateAccount: true,
					PinCode:       pinCode,
					ChannelId:     channelId,
					Username:      ctx.GetUserName(),
					Nickname:      ctx.GetNickname(),
				})
				if err != nil {
					log.Error().Fields(map[string]interface{}{
						"action": "get user",
						"error":  err.Error(),
						"openId": ctx.Requester.RequesterOpenId,
					}).Send()
					return he.NewServerError(pconst.CodeWalletRequestError, "", err)
				} else if addUserResp.CommonResponse.Code != he.Success {

					return tcontext.RespToError(addUserResp.CommonResponse)
				} else {
					log.Debug().Fields(map[string]interface{}{"action": "init user", "userNo": addUserResp.Data.UserNo, "username": ctx.GetAvailableName(), "address": addUserResp.Data.DefaultAccountAddr, "pinCode": pinCode}).Send()
					log.Info().Fields(map[string]interface{}{"action": "init user", "userNo": addUserResp.Data.UserNo, "username": ctx.GetAvailableName(), "address": addUserResp.Data.DefaultAccountAddr}).Send()
					isCreateAccount = true
					user = addUserResp.Data
				}
			} else {
				// todo switch to TSTORE
				UserChannel[ctx.OpenId()] = ctx.Requester.RequesterChannelId
				// ignore send remind message error for create account
				if _, herr := ctx.Send(ctx.U.SentFrom().ID, text.ClickStart, nil, false, false); herr != nil {
				}
			}
		} else {
			return tcontext.RespToError(getUserResp.CommonResponse)
		}
	} else {
		user = getUserResp.Data
	}

	//cmdDesc := "⚙️ Commands\n"
	//for _, v := range GetCmdList() {
	//	cmdDesc += fmt.Sprintf("/%s %s\n", v, Desc[v])
	//}
	//content := "ℹ️ User Guide\n"
	//content += text.Introduce
	//
	//content += "\n"
	//content += "\n"
	//content += cmdDesc
	//content += "\n"

	if ctx.U.Message.Chat.IsPrivate() {

		if user == nil {
			return he.NewBusinessError(pconst.CodeUserNotInit, "", nil)
		}

		walletContent := "⚡️ Wallet\n"
		if isCreateAccount {
			walletContent += fmt.Sprintf(text.CreateAccountSuccess, user.DefaultAccountAddr, pinCode)
		} else {
			walletContent += fmt.Sprintf(text.GetAccountSuccess, user.DefaultAccountAddr)
		}

		if text.CustomStartMenu != "" {
			walletContent = fmt.Sprintf("%s\n%s", text.CustomStartMenu, walletContent)
		}

		//if _, herr := ctx.Send(ctx.U.SentFrom().ID, content, nil, false); herr != nil {
		//	return herr
		//}

		if _, herr := ctx.Send(ctx.U.SentFrom().ID, walletContent, nil, true, false); herr != nil {
			return herr
		}
	} else {
		groupContent := fmt.Sprintf(text.SwitchPrivate, ctx.GetNickNameMDV2())

		if text.CustomStartMenu != "" {
			groupContent = fmt.Sprintf("%s\n%s", text.CustomStartMenu, groupContent)
		}

		var inlineKeyboard *tgbotapi.InlineKeyboardMarkup
		var deadlineTime time.Duration
		if user == nil {
			inlineKeyboard, deadlineTime = inline_keybord.NewForwardCreateKeyBoard(ctx)
		} else {
			inlineKeyboard, deadlineTime = inline_keybord.NewForwardPrivateKeyBoard(ctx)
		}

		replyMsg, herr := ctx.Reply(ctx.U.FromChat().ID, groupContent, inlineKeyboard, true)
		if herr != nil {
			return herr
		}
		inline_keybord.DeleteDeadKeyboard(ctx, deadlineTime, &replyMsg)
	}
	return nil

}
