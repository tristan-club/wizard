package cmd_open_envelope

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/log"
	"github.com/tristan-club/wizard/pkg/mdparse"
	"github.com/tristan-club/wizard/pkg/tstore"
	"strconv"
	"strings"
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

func openEnvelopeHandler(ctx *tcontext.Context) error {
	params := strings.Split(ctx.U.CallbackData(), "/")
	if len(params) != 2 {
		log.Error().Fields(map[string]interface{}{"action": "invalid envelope params", "payload": ctx.U.CallbackData()}).Send()
		return he.NewServerError(he.CodeInvalidPayload, "", fmt.Errorf("invalid payload"))
	}
	envelopeId, err := strconv.ParseInt(params[1], 10, 32)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "invalid envelope id", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeInvalidPayload, "", err)
	}

	openEnvelopeResp, err := ctx.CM.OpenEnvelope(ctx.Context, &controller_pb.OpenEnvelopeReq{
		Address:    ctx.Requester.RequesterDefaultAddress,
		EnvelopeNo: "",
		EnvelopeId: uint32(envelopeId),
		IsWait:     false,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	}

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
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if openEnvelopeResp.CommonResponse.Code != he.Success {
		if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_OPENED || openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_SOLD_OUT {
			if _, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(mdparse.ParseV2(text.BusinessError), ctx.GetNickNameMDV2(), "open envelope command", mdparse.ParseV2(openEnvelopeResp.CommonResponse.Message)), nil, true, false); herr != nil {
				return herr
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
	pendingMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.OpenEnvelopeTransactionProcessing, ctx.GetNickNameMDV2(), envelopeId, amountLabel, assetSymbol, fmt.Sprintf("%s%s", pconst.GetExplore(chainType, pconst.ExploreTypeTx), openEnvelopeResp.Data.TxHash)), nil, true, false)
	if herr != nil {
		return herr
	}

	getDataResp, err := ctx.CM.GetTx(context.Background(), &controller_pb.GetTxReq{TxHash: openEnvelopeResp.Data.TxHash})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if getDataResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(getDataResp.CommonResponse)
	}

	ctx.TryDeleteMessage(pendingMsg)
	if _, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.OpenEnvelopeSuccess, ctx.GetNickNameMDV2(), envelopeId, mdparse.ParseV2(amount), mdparse.ParseV2(assetSymbol), mdparse.ParseV2(fmt.Sprintf("%s%s", pconst.GetExplore(chainType, pconst.ExploreTypeTx), openEnvelopeResp.Data.TxHash))), nil, true, false); herr != nil {
		return herr
	}
	return nil
}
