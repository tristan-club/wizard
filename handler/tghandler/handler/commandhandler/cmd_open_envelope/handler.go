package cmd_open_envelope

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/bignum"
	"github.com/tristan-club/kit/chain_info"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/mdparse"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/tstore"
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

var Handler = chain.NewChainHandler(cmd.CmdOpenEnvelope, openEnvelopeHandler).
	AddCmdParser(func(u *tgbotapi.Update) string {
		if strings.HasPrefix(u.CallbackData(), cmd.CmdOpenEnvelope) {
			return cmd.CmdOpenEnvelope
		}
		return ""
	}).
	AddPreHandler(prehandler.OnlyPublic).
	AddPreHandler(prehandler.SetFrom)

func IsOpenEnvelopeCmd(text string) bool {
	return strings.HasPrefix(text, cmd.CmdOpenEnvelope)
}

func IsBridgeCmd(text string) bool {
	return strings.HasPrefix(text, cmd.CmdBridge)
}

// todo 检查为啥红包发出来两个人名一样
func openEnvelopeHandler(ctx *tcontext.Context) error {

	uid := util.GenerateUuid(true)

	log.Info().Msgf("user %s start opening envelope, uid %s", ctx.GetUserName(), uid)
	params := strings.Split(ctx.U.CallbackData(), "/")
	if len(params) != 2 {
		log.Error().Fields(map[string]interface{}{"action": "invalid envelope params", "payload": ctx.U.CallbackData()}).Send()
		return he.NewServerError(pconst.CodeInvalidPayload, "", fmt.Errorf("invalid payload"))
	}
	envelopeId, err := strconv.ParseInt(params[1], 10, 32)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "invalid envelope id", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeInvalidPayload, "", err)
	}

	openEnvelopeResp, err := ctx.CM.OpenEnvelope(ctx.Context, &controller_pb.OpenEnvelopeReq{
		Address:    ctx.Requester.RequesterDefaultAddress,
		EnvelopeNo: "",
		EnvelopeId: uint32(envelopeId),
		IsWait:     true,
		ChannelId:  ctx.Requester.RequesterChannelId,
		ReceiverNo: ctx.Requester.RequesterUserNo,
	})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	}

	log.Info().Msgf("user %s send open envelope request, uid %s", ctx.GetUserName(), uid)
	//delete envelope keyboard if it is sold out or invalid
	defer func() {

		if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_SOLD_OUT || openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOP_STATE_INVALID {
			messageIdStr, err := tstore.PBGetStr(fmt.Sprintf("%s%d", pconst.EnvelopeStorePrefix, envelopeId), pconst.EnvelopeStorePath)
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "get envelope message error", "error": err.Error()}).Send()
				return
			}
			messageId, err := strconv.ParseInt(messageIdStr, 10, 64)
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "parse message id error", "error": err.Error(), "raw": messageIdStr}).Send()
				return
			}
			ctx.DeleteMessage(ctx.U.FromChat().ID, int(messageId))

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
			return tcontext.RespToError(openEnvelopeResp.CommonResponse)
		}

	}

	amount := openEnvelopeResp.Data.Amount
	assetSymbol := openEnvelopeResp.Data.AssetSymbol
	chainType := openEnvelopeResp.Data.ChainType
	amountLabel := strings.ReplaceAll(amount, ".", "\\.")
	pendingMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.OpenEnvelopeTransactionProcessing, ctx.GetNickNameMDV2(), envelopeId, amountLabel, assetSymbol), nil, true, false)
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
	envelopeMsgId, err := tstore.PBGetStr(fmt.Sprintf("%s%d", pconst.EnvelopeStorePrefix, envelopeId), pconst.EnvelopeStorePath)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "get envelope msg id error", "error": err.Error()}).Send()
	} else {

		getEnvelopeResp, err := ctx.CM.GetEnvelope(ctx.Context, &controller_pb.GetEnvelopeReq{EnvelopeId: uint32(envelopeId), WithClaimList: true, WaitSuccess: true})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "call wallet", "error": err.Error()}).Send()
		} else if getEnvelopeResp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "get envelope", "error": getEnvelopeResp.CommonResponse}).Send()
		} else {

			envelopeDetail := fmt.Sprintf(text.EnvelopeDetail,
				ctx.GenerateNickName(mdparse.ParseV2(getEnvelopeResp.Data.CreatorName), strconv.FormatInt(getEnvelopeResp.Data.CreatorOpenId, 10)),
				mdparse.ParseV2(bignum.CutDecimal(new(big.Int).SetUint64(getEnvelopeResp.Data.RemainAmount), 4, 4)), mdparse.ParseV2(getEnvelopeResp.Data.AssetSymbol),
				getEnvelopeResp.Data.Quantity-getEnvelopeResp.Data.RemainQuantity, getEnvelopeResp.Data.Quantity)

			var claimHistory string
			for _, claim := range getEnvelopeResp.Data.ClaimList {
				labelLen := 20
				labelName := make([]rune, labelLen)
				nicknameRune := []rune(claim.ReceiverNickname)
				for i := 0; i < len(nicknameRune) && i < labelLen; i++ {
					labelName[i] = nicknameRune[i]
				}

				claimHistory += fmt.Sprintf("%s   %s %s   TXN URL\\: [click to view](%s)\n\n",
					ctx.GenerateNickName(mdparse.ParseV2(string(labelName)), claim.ReceiverOpenId),
					mdparse.ParseV2(claim.Amount),
					mdparse.ParseV2(claim.AssetSymbol), mdparse.ParseV2(chain_info.GetExplorerTargetUrl(getEnvelopeResp.Data.ChainId, claim.TxHash, chain_info.ExplorerTargetTransaction)))
			}
			envelopeDetail = envelopeDetail + "\n\n🎊Claim History:\n\n" + claimHistory
			msgId, _ := strconv.ParseInt(envelopeMsgId, 10, 64)
			openButton := tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text.OpenEnvelope, fmt.Sprintf("%s/%d", cmd.CmdOpenEnvelope, envelopeId))},
			)
			herr = ctx.EditMessageAndKeyboard(ctx.U.FromChat().ID, int(msgId), envelopeDetail, &openButton, true, true)
			if herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "edit red envelope error", "error": herr.Error()}).Send()
			}

		}
	}

	log.Info().Msgf("user %s get open envelope tx hash uid %s", ctx.GetUserName(), uid)

	if openMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.OpenEnvelopeSuccess, ctx.GetNickNameMDV2(), envelopeId, mdparse.ParseV2(amount), mdparse.ParseV2(assetSymbol), mdparse.ParseV2(pconst.GetExplore(chainType, getDataResp.Data.TxHash, chain_info.ExplorerTargetTransaction))), nil, true, true); herr != nil {
		return herr
	} else {
		ctx.SetDeadlineMsg(openMsg.Chat.ID, openMsg.MessageID, pconst.GroupMentionDeadline)
	}

	return nil
}
