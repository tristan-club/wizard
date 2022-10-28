package cmd_start

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/inline_keybord"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/dingding"
	"github.com/tristan-club/wizard/pkg/tstore"
	"strings"
	"time"
)

type OpenEnvelopePayload struct {
}

var Handler = chain.NewChainHandler(cmd.CmdStart, startSendHandler)

var UserChannel = map[string]string{}

func startSendHandler(ctx *tcontext.Context) error {

	var user *controller_pb.User
	var isCreateUser bool
	var pinCode string
	var isStartBot bool
	var activityId string
	var inviteeId string
	var inviteGroupId string

	if len(ctx.CmdParam) != 0 {
		log.Info().Fields(map[string]interface{}{"action": "get start param", "param": ctx.CmdParam}).Send()
	}

	if len(ctx.CmdParam) == 1 {
		if ctx.CmdParam[0] == pconst.DefaultDeepLinkStart {
			log.Info().Msgf("openId %s use start deep link", ctx.OpenId())
		} else {
			inviteArray := strings.Split(ctx.CmdParam[0], "_")
			if len(inviteArray) != 3 {
				robot := dingding.NewRobot(config.GetDingDingToken(), "", "", "")
				robot.SendMarkdownMessage("## Telegram Wizard", fmt.Sprintf("invalid cmd param for /start, openId %s, param %s", ctx.OpenId(), ctx.CmdParam[0]), nil, false)
				log.Error().Fields(map[string]interface{}{"action": "invalid cmd param for  start", "openId": ctx.OpenId(), "param": ctx.CmdParam, "ctx": ctx}).Send()
			} else {
				isStartBot = true
				activityId = inviteArray[0]
				inviteeId = inviteArray[1]
				inviteGroupId = inviteArray[2]
			}
		}
	}

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
	}

	if getUserResp.CommonResponse.Code != he.Success {
		if getUserResp.CommonResponse.Code == pconst.CODE_USER_NOT_EXIST || getUserResp.Data.DefaultAccountAddr == "" {
			if ctx.U.FromChat().IsPrivate() {

				channelId := ctx.Requester.GetRequesterChannelId()
				if channelId == "" {
					channelId = UserChannel[ctx.OpenId()]
				}

				addUserReq := &controller_pb.AddUserReq{
					OpenId:        ctx.Requester.RequesterOpenId,
					OpenType:      int32(ctx.Requester.RequesterOpenType),
					IsOpenInit:    true,
					CreateAccount: false,
					ChannelId:     channelId,
					Username:      ctx.GetUserName(),
					Nickname:      ctx.GetNickname(),
					AppId:         ctx.Requester.RequesterAppId,
				}

				if !isStartBot {
					pinCode = pconst.DefaultPinCode
					addUserReq.PinCode = pinCode
					addUserReq.CreateAccount = true
				}

				addUserResp, err := ctx.CM.AddUser(ctx.Context, addUserReq)
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
					isCreateUser = true
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

	if ctx.U.Message.Chat.IsPrivate() {

		if user == nil {
			return he.NewBusinessError(pconst.CodeUserNotInit, "", nil)
		}

		if isStartBot {
			robot := dingding.NewRobot(config.GetDingDingToken(), "", "", "")
			inviteLink, err := tstore.PBGetStr("tg_invite", inviteGroupId)
			if err != nil {
				robot.SendMarkdownMessage("## Telegram Wizard", fmt.Sprintf("get invaite link error %s , groupId %s, activityId %s, inviteeId %s", err.Error(), inviteGroupId, activityId, inviteeId), nil, false)
				log.Error().Fields(map[string]interface{}{"action": "get invite link error", "error": err.Error(), "ctx": ctx}).Send()
				return he.NewServerError(he.ServerError, "", err)
			} else if inviteLink == "" {
				robot.SendMarkdownMessage("## Telegram Wizard", fmt.Sprintf("get empty invaite link , groupId %s, activityId %s, inviteeId %s", inviteGroupId, activityId, inviteeId), nil, false)
				log.Info().Msgf("emptyInviteLink, groupId %s, activityId %s, inviteeId %s", inviteGroupId, activityId, inviteeId)
				return he.NewServerError(he.ServerError, "invalid invite link", nil)
			}

			log.Info().Msgf("got invite link %s, groupId %s, activityId %s, inviteeId %s", inviteLink, inviteGroupId, activityId, inviteeId)

			inviteContent := text.StartInviteText
			if inviteContent == "" {
				inviteContent = fmt.Sprintf(text.StartBotDefaultText, ctx.GetNickNameMDV2())
			}

			if err = tstore.PBSaveString(fmt.Sprintf("task-%s", inviteGroupId), ctx.OpenId(), fmt.Sprintf("%s_%s", activityId, inviteeId)); err != nil {
				log.Error().Fields(map[string]interface{}{"action": "save task invite error", "error": err.Error(), "ctx": ctx}).Send()
				return he.NewServerError(he.ServerError, "", err)
			}

			forwardIkm := tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL(text.ButtonJoin, inviteLink)})
			if msg, herr := ctx.Send(ctx.U.SentFrom().ID, inviteContent, forwardIkm, true, false); herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send forward ikm error", "error": herr.Error()}).Send()
				return herr
			} else {
				inline_keybord.DeleteDeadKeyboard(ctx, pconst.COMMON_KEYBOARD_DEADLINE, msg)
			}

		} else {

			suffix := fmt.Sprintf("?app_id=%s", ctx.Requester.RequesterAppId)

			if config.UseTemporaryToken() {

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
					suffix += fmt.Sprintf("&token=%s", initTemporaryTokenResp.Data.Token)
				}

			}

			suffix = strings.ReplaceAll(suffix, " ", "%20")

			log.Debug().Msgf("temporary print url: %s", suffix)
			//ikm := tgbotapi.NewInlineKeyboardMarkup(
			//	[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text.ButtonHelp, cmd.CmdMenu), tgbotapi.NewInlineKeyboardButtonData(text.ChangePinCode, cmd.CmdChangePinCode),
			//		tgbotapi.NewInlineKeyboardButtonData(text.SubmitMetamask, cmd.CmdSubmitMetamask)},
			//)
			_ = ctx.SetChatMenuButton(&tgbotapi.MenuButton{
				Type:   "web_app",
				Text:   "Account",
				WebApp: &tgbotapi.WebAppInfo{URL: fmt.Sprintf("%s%s", pconst.WebAppMenuUrl, suffix)},
			})

			line1 := []tgbotapi.KeyboardButton{{
				Text:   text.KBAccount,
				WebApp: &tgbotapi.WebAppInfo{URL: fmt.Sprintf("%s%s", pconst.WebAppMenuUrl, suffix)},
			}, {
				Text:   text.KBProfile,
				WebApp: &tgbotapi.WebAppInfo{URL: fmt.Sprintf("%s%s", pconst.WebAppProfileUrl, suffix)},
			}}
			line2 := []tgbotapi.KeyboardButton{{
				Text:   text.KBActivity,
				WebApp: &tgbotapi.WebAppInfo{URL: fmt.Sprintf("%s%s", pconst.WebAppActivityUrl, suffix)},
			}, {
				Text: text.KBChangePinCode,
			}}
			line3 := []tgbotapi.KeyboardButton{{
				Text: text.KBBalance,
			}, {
				Text: text.KBHelp,
			}}

			keyboardBt := &tgbotapi.ReplyKeyboardMarkup{
				Keyboard: [][]tgbotapi.KeyboardButton{line1, line2, line3},
			}

			walletContent := "⚡️ Wallet\n"
			if isCreateUser {
				walletContent += fmt.Sprintf(text.CreateAccountSuccess, user.DefaultAccountAddr, pinCode)
			} else {
				walletContent += fmt.Sprintf(text.GetAccountSuccess, user.DefaultAccountAddr)
			}

			//if text.CustomStartMenu != "" {
			//	walletContent = fmt.Sprintf("%s\n%s", text.CustomStartMenu, walletContent)
			//}

			//if _, herr := ctx.Send(ctx.U.SentFrom().ID, content, nil, false); herr != nil {
			//	return herr
			//}

			_, herr := ctx.Send(ctx.U.SentFrom().ID, walletContent, keyboardBt, true, false)
			if herr != nil {

				log.Error().Fields(map[string]interface{}{"action": "send wallet content error", "error": herr.Error(), "ctx": ctx}).Send()

				_, herr = ctx.Send(ctx.U.SentFrom().ID, walletContent, nil, true, false)
				if herr != nil {
					log.Error().Fields(map[string]interface{}{"action": "register keyboard bt error", "error": herr.Error(), "ikm": keyboardBt}).Send()
					if isCreateUser {
						req := &controller_pb.ChangeAccountPinCodeReq{
							Address:    user.DefaultAccountAddr,
							OldPinCode: pinCode,
							NewPinCode: pconst.DefaultPinCode,
						}
						updateUserResp, err := ctx.CM.ChangeAccountPinCode(ctx.Context, req)
						if err != nil {
							log.Error().Fields(map[string]interface{}{"action": "request controller svc error", "error": err.Error(), "req": req}).Send()
						} else if updateUserResp.CommonResponse.Code != he.Success {
							log.Error().Fields(map[string]interface{}{"action": "update account pin code error", "error": updateUserResp, "req": req}).Send()
						}
					}
				}

				return herr
			}

			if isCreateUser {

				_, herr = ctx.SendPhoto(ctx.U.SentFrom().ID, "", nil, true, pconst.UserGuideImgUrl)

				go func() {
					time.Sleep(time.Minute * 5)
					ikm := tgbotapi.NewInlineKeyboardMarkup(
						[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text.ChangePinCode, cmd.CmdChangePinCode)},
					)
					if _, herr = ctx.SendPhoto(ctx.U.SentFrom().ID, text.ChangeYourPin, ikm, false, pconst.ChangePinCodeImgUrl); herr != nil {
						log.Error().Fields(map[string]interface{}{"action": "send change pin img error", "error": herr.Error()}).Send()
					}
				}()
			}

			//if isCreateUser {
			//	go func() {
			//		time.Sleep(time.Second * 5)
			//		if _, herr := ctx.Send(ctx.U.SentFrom().ID, text.RecommendChangePinCode, nil, true, false); herr != nil {
			//			log.Error().Fields(map[string]interface{}{"action": "send msg error", "error": herr}).Send()
			//		}
			//	}()
			//}
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
