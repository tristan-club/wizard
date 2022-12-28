package cmd_envelope

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/bignum"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/mdparse"
	"github.com/tristan-club/kit/tstore"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/envelope_limiter"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_start"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/util"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type OpenEnvelopePayload struct {
	ChainType   uint32 `json:"chain_type"`
	EnvelopeId  uint32 `json:"envelope_id"`
	ChannelId   string `json:"channel_id"`
	AssetSymbol string `json:"asset_symbol"`
}

func NoAddressUserHandler(ctx *tcontext.Context) error {

	if ctx.U.CallbackData() == "" {
		log.Error().Fields(map[string]interface{}{"action": "invalid envelope config", "ctx": ctx}).Send()
		return he.NewServerError(he.ServerError, "", fmt.Errorf("invlaid envelope config"))
	}

	cid, _ := customid.ParseCustomId(ctx.U.CallbackData())

	openButton := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL(text.ClaimAndStart, ctx.GenerateDeepLink(cid))},
	)

	if replyMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.NoAddressEnvelopeUser, ctx.GetMentionName()), &openButton, false, true); herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg error", "error": herr}).Send()
		return herr
	} else {
		ctx.SetDeadlineMsg(replyMsg.Chat.ID, replyMsg.MessageID, pconst.GroupMentionDeadline)
	}
	return nil
}

var OpenEnvelopeHandler = chain.NewChainHandler(cmd.CmdOpenEnvelope, openEnvelopeHandler).
	AddCmdParser(func(u *tgbotapi.Update) string {
		// 有钱包用户通过按钮领取
		var cid *customid.CustomId
		var ok bool
		if u.Message != nil {
			if u.Message.IsCommand() && u.Message.Command() == cmd.CmdStart {
				cid, ok = customid.ParseCustomId(u.Message.CommandArguments())
			}
			// 没钱包的用户通过deeplink跳转，创建钱包并领取
		} else if u.CallbackData() != "" {
			cid, ok = customid.ParseCustomId(u.CallbackData())
		}

		if !ok || cid == nil || cid.GetCustomType() != pconst.CustomIdOpenEnvelope {
			return ""
		}

		return cmd.CmdOpenEnvelope
	}).
	AddPreHandler(prehandler.SetFrom)

// todo 检查为啥红包发出来两个人名一样
func openEnvelopeHandler(ctx *tcontext.Context) error {

	var userHasAddress bool
	var envelopeNo string
	var option controller_pb.ENVELOPE_OPTION
	var channelId int64
	var channelUsername string
	var fromId string
	var address string

	var cid *customid.CustomId

	if ctx.U.Message != nil {
		cid, _ = customid.ParseCustomId(ctx.U.Message.CommandArguments())
	} else if cbd := ctx.U.CallbackData(); cbd != "" {
		cid, _ = customid.ParseCustomId(ctx.U.CallbackData())
		userHasAddress = true
		channelId = ctx.U.FromChat().ID
		channelUsername = ctx.U.FromChat().UserName
		fromId = ctx.Requester.RequesterUserNo
		address = ctx.Requester.RequesterDefaultAddress
	} else {
		log.Error().Fields(map[string]interface{}{"action": "invalid envelope parser", "ctx": ctx}).Send()
		return fmt.Errorf("invalid envelope parser")
	}

	envelopeNo = cid.GetId()
	option = controller_pb.ENVELOPE_OPTION(cid.GetCallbackType())
	if channelId == 0 {
		chId, err := tstore.PBGetStr(fmt.Sprintf("%s%s", pconst.EnvelopeStorePrefix, envelopeNo), pconst.EnvelopeStorePathChannelId)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "get envelope ch id error", "error": err.Error(), "ctx": ctx}).Send()
			return he.NewServerError(he.ServerError, "parse envelope info error", err)
		}
		a := strings.Split(chId, "/")
		if len(a) != 2 {
			log.Error().Fields(map[string]interface{}{"action": "invalid envelope payload", "ctx": ctx}).Send()
			return he.NewServerError(he.ServerError, "", fmt.Errorf("invalid envelope payload"))
		}
		channelId, _ = strconv.ParseInt(a[0], 10, 64)
		channelUsername = a[1]
	}

	if !userHasAddress {
		ctx.Payload = cmd_start.StartParam{IgnoreGuideMsg: true}
		ctx.CmdParam = []string{pconst.DefaultDeepLinkStart}
		err := cmd_start.Handler.Handle(ctx)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "handle start error", "error": err.Error(), "ctx": ctx}).Send()
			return err
		}
		result, ok := ctx.Result.(*cmd_start.StartResult)
		if !ok {
			log.Error().Fields(map[string]interface{}{"action": "invalid start result", "ctx": ctx}).Send()
			return he.NewServerError(he.ServerError, "", fmt.Errorf("envelope config error"))
		}
		fromId = result.UserId
		address = result.Address
	}

	uid := util.GenerateUuid(true)
	log.Info().Msgf("user %s start opening envelope, uid %s", ctx.GetUserName(), uid)

	if envelope_limiter.CheckEnvelopeClaim(envelopeNo, fromId) {
		log.Debug().Fields(map[string]interface{}{"action": "dup claim", "userId": fromId}).Send()
		return nil
	}

	openEnvelopeResp, err := ctx.CM.OpenEnvelope(ctx.Context, &controller_pb.OpenEnvelopeReq{
		AppId:          ctx.Requester.RequesterAppId,
		Address:        address,
		EnvelopeNo:     envelopeNo,
		IsWait:         true,
		ReceiverNo:     fromId,
		EnvelopeOption: controller_pb.ENVELOPE_OPTION(option),
	})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "open envelope error", "error": err.Error(), "ctx": ctx}).Send()
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	}

	log.Info().Msgf("user %s send open envelope request, uid %s", ctx.GetUserName(), uid)
	//delete envelope keyboard if it is sold out or invalid
	defer func() {

		if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_SOLD_OUT || openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOP_STATE_INVALID {
		}
	}()

	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if openEnvelopeResp.CommonResponse.Code != he.Success {
		if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_OPENED || openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_SOLD_OUT {
			if errMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(mdparse.ParseV2(text.BusinessError), ctx.GetNickNameMDV2(), "open envelope command", mdparse.ParseV2(openEnvelopeResp.CommonResponse.Message)), nil, true, false); herr != nil {
				return herr
			} else {
				ctx.SetDeadlineMsg(errMsg.Chat.ID, errMsg.MessageID, pconst.GroupMentionDeadline)
			}
			return nil
		} else if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOP_STATE_INVALID {
			// todo
			return tcontext.RespToError(openEnvelopeResp.CommonResponse)
		} else {
			t := openEnvelopeResp.CommonResponse.Inner
			if t == "" {
				t = openEnvelopeResp.CommonResponse.Message
			}
			return he.NewBusinessError(he.ServerError, fmt.Sprintf("Operation Failed\n%s", t), nil)
		}

	}

	amount := openEnvelopeResp.Data.Amount
	assetSymbol := openEnvelopeResp.Data.AssetSymbol
	chainType := openEnvelopeResp.Data.ChainType
	amountLabel := strings.ReplaceAll(amount, ".", "\\.")
	pendingMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.OpenEnvelopeTransactionProcessing, ctx.GetNickNameMDV2(), envelopeNo, amountLabel, assetSymbol), nil, true, false)
	if herr != nil {
		return herr
	}
	log.Info().Msgf("user %s send open envelope pending msg uid %s", ctx.GetUserName(), uid)

	time.Sleep(time.Second * 5)

	getDataResp, err := ctx.CM.GetTx(context.Background(), &controller_pb.GetTxReq{TxId: openEnvelopeResp.Data.TxId, IsWait: true})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if getDataResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(getDataResp.CommonResponse)
	}
	ctx.TryDeleteMessage(pendingMsg)
	envelopeMsgId, err := tstore.PBGetStr(fmt.Sprintf("%s%s", pconst.EnvelopeStorePrefix, envelopeNo), pconst.EnvelopeStorePathMsgId)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "get envelope msg id error", "error": err.Error()}).Send()
	} else {

		getEnvelopeResp, err := ctx.CM.GetEnvelope(ctx.Context, &controller_pb.GetEnvelopeReq{EnvelopeNo: envelopeNo, WithClaimList: true, WaitSuccess: true})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "call wallet", "error": err.Error()}).Send()
		} else if getEnvelopeResp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "get envelope", "error": getEnvelopeResp.CommonResponse}).Send()
		} else {
			title := text.EnvelopeTitleOrdinary
			if controller_pb.ENVELOPE_OPTION(option) == controller_pb.ENVELOPE_OPTION_HAS_CAT {
				title = text.EnvelopeTitleCAT
			}
			title = fmt.Sprintf(title, envelopeNo, ctx.GenerateNickName(mdparse.ParseV2(getEnvelopeResp.Data.CreatorName), strconv.FormatInt(getEnvelopeResp.Data.CreatorOpenId, 10)))
			title += "\n\n"
			envelopeDetail := fmt.Sprintf(text.EnvelopeDetail,
				mdparse.ParseV2(bignum.CutDecimal(new(big.Int).SetUint64(getEnvelopeResp.Data.RemainAmount), 4, 4)), mdparse.ParseV2(getEnvelopeResp.Data.AssetSymbol),
				getEnvelopeResp.Data.Quantity-getEnvelopeResp.Data.RemainQuantity, getEnvelopeResp.Data.Quantity)

			envelopeDetail = title + envelopeDetail

			var claimHistory string
			for _, claim := range getEnvelopeResp.Data.ClaimList {
				labelLen := 20
				labelName := make([]rune, labelLen)
				nicknameRune := []rune(claim.ReceiverNickname)
				for i := 0; i < len(nicknameRune) && i < labelLen; i++ {
					labelName[i] = nicknameRune[i]
				}

				// todo 没有txhash，用address代替
				var txnUrl string
				if claim.TxHash != "" {
					txnUrl = chain_info.GetExplorerTargetUrl(getEnvelopeResp.Data.ChainId, claim.TxHash, chain_info.ExplorerTargetTransaction)
				} else {
					txnUrl = chain_info.GetExplorerTargetUrl(getEnvelopeResp.Data.ChainId, claim.ReceiverAddress, chain_info.ExplorerTargetAddress)
				}

				claimHistory += fmt.Sprintf("%s   %s %s   TXN URL\\: [click to view](%s)\n\n",
					ctx.GenerateNickName(mdparse.ParseV2(string(labelName)), claim.ReceiverOpenId),
					mdparse.ParseV2(claim.Amount),
					mdparse.ParseV2(claim.AssetSymbol), mdparse.ParseV2(txnUrl))
			}
			envelopeDetail = envelopeDetail + "\n\n🎊Claim History:\n\n" + claimHistory
			msgId, _ := strconv.ParseInt(envelopeMsgId, 10, 64)
			var openButton *tgbotapi.InlineKeyboardMarkup

			if getEnvelopeResp.Data.RemainQuantity != 0 {
				openButton = &tgbotapi.InlineKeyboardMarkup{}
				*openButton = tgbotapi.NewInlineKeyboardMarkup(
					[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text.OpenEnvelope, cid.String())},
				)
			}
			herr = ctx.EditMessageAndKeyboard(channelId, int(msgId), envelopeDetail, openButton, true, true)
			if herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "edit red envelope error", "error": herr.Error()}).Send()
			}

		}
	}

	log.Info().Msgf("user %s get open envelope tx hash uid %s", ctx.GetUserName(), uid)

	if ctx.U.FromChat().IsPrivate() {
		if _, herr = ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.OpenEnvelopeSuccess, ctx.GetNickNameMDV2(), envelopeNo,
			mdparse.ParseV2(amount), mdparse.ParseV2(assetSymbol), mdparse.ParseV2(pconst.GetExplore(chainType, getDataResp.Data.TxHash, chain_info.ExplorerTargetTransaction)),
		), nil, true, true); herr != nil {
			return herr
		}
	}

	if _, herr = ctx.Send(channelId, fmt.Sprintf(text.OpenEnvelopeSuccessGroupMsg, ctx.GetNickNameMDV2(), envelopeNo,
		mdparse.ParseV2(amount), mdparse.ParseV2(assetSymbol), mdparse.ParseV2(pconst.GetExplore(chainType, getDataResp.Data.TxHash, chain_info.ExplorerTargetTransaction)),
		fmt.Sprintf("%s/%s/%s", pconst.TGLink, channelUsername, envelopeMsgId)), nil, true, true); herr != nil {
		return herr
	}

	//if sendGuideMsg {
	//	_, herr = ctx.SendPhoto(ctx.U.SentFrom().ID, "", nil, true, pconst.UserGuideImgUrl)
	//}

	return nil
}
