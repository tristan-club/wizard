package cmd_create_envelope

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/mdparse"
	"github.com/tristan-club/wizard/pkg/tstore"
	"strconv"
	"strings"
)

const (
	EnvelopeNumMin = 1
	EnvelopeNumMax = 20

	AmountMin = "0.0001"
	AmountMax = "10000000000"
)

var envelopeTypeText = []string{"Average Amount", "Random Amount"}
var envelopeTypeValue = []int64{1, 2}

type CreateEnvelopePayload struct {
	UserNo       string `json:"user_no"`
	From         string `json:"from"`
	ChainType    uint32 `json:"chain_type"`
	Asset        string `json:"asset"`
	AssetSymbol  string `json:"asset_symbol"`
	EnvelopeType uint32 `json:"envelope_type"`
	ChannelId    string `json:"channel_id"`
	Quantity     uint64 `json:"quantity"`
	Amount       string `json:"amount"`
	PinCode      string `json:"pin_code"`
}

var Handler *chain.ChainHandler

var enterEnvelopeTypeNode *chain.Node

func init() {
	enterEnvelopeTypeNode = new(chain.Node)
	*enterEnvelopeTypeNode = *presetnode.EnterTypeNode

	enterEnvelopeQuantityNode := chain.NewNode(presetnode.AskForQuantity, prechecker.MustBeMessage, presetnode.EnterQuantity)

	Handler = chain.NewChainHandler(cmd.CmdCreateEnvelope, createEnvelopeSendHandler).
		AddPreHandler(prehandler.OnlyPublic).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.SelectChainNode, nil).
		AddPresetNode(presetnode.EnterAssetNode, nil).
		AddPresetNode(enterEnvelopeTypeNode, &presetnode.EnterTypeParam{
			ChoiceText:  envelopeTypeText,
			ChoiceValue: envelopeTypeValue,
			Content:     text.SelectEnvelopeType,
			ParamKey:    "envelope_type",
		}).
		AddPresetNode(presetnode.EnterAmountNode, &presetnode.AmountParam{
			Min:          AmountMin,
			Max:          AmountMax,
			Content:      text.EnterAmountWithRange,
			ParamKey:     "amount",
			CheckBalance: true,
		}).
		AddPresetNode(enterEnvelopeQuantityNode, &presetnode.EnterQuantityParam{
			Min:      EnvelopeNumMin,
			Max:      EnvelopeNumMax,
			Content:  text.EnterEnvelopeQuantity,
			ParamKey: "quantity",
		}).
		AddPresetNode(presetnode.EnterPinCodeHandler, nil)
}

func createEnvelopeSendHandler(ctx *tcontext.Context) error {

	var payload = &CreateEnvelopePayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	channelId, err := strconv.ParseInt(payload.ChannelId, 10, 64)
	if err != nil {
		return he.NewServerError(pconst.CodeInvalidPayload, "", err)
	}

	tokenType := pconst.TokenTypeInternal
	if payload.Asset != "" && payload.Asset != "INTERNAL" && strings.HasPrefix(payload.Asset, "0x") {
		tokenType = pconst.TokenTypeErc20
	}

	createEnvelopeReq := &controller_pb.AddEnvelopeReq{
		FromId:          payload.UserNo,
		ChainType:       payload.ChainType,
		ChannelId:       payload.ChannelId,
		ChainId:         pconst.GetChainId(payload.ChainType),
		TokenType:       uint32(tokenType),
		Address:         payload.From,
		ContractAddress: payload.Asset,
		Amount:          payload.Amount,
		Quantity:        payload.Quantity,
		EnvelopeType:    payload.EnvelopeType,
		Blessing:        "",
		PinCode:         payload.PinCode,
		IsWait:          false,
	}

	createRedEnvelope, err := ctx.CM.AddEnvelope(ctx.Context, createEnvelopeReq)
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if createRedEnvelope.CommonResponse.Code != he.Success {
		return tcontext.RespToError(createRedEnvelope.CommonResponse)
	}

	pendingMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.EnvelopePreparing, fmt.Sprintf("%s%s", pconst.GetExplore(payload.ChainType, pconst.ExploreTypeAddress), createRedEnvelope.Data.AccountAddress)), nil, true, false)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "send pending tx error", "error": herr.Error()}).Send()
		return herr
	}

	requesterCtx, herr := ctx.CopyRequester()
	if herr != nil {
		return herr
	}

	//time.Sleep(time.Second * 1)
	envelopeResp, err := ctx.CM.GetEnvelope(requesterCtx, &controller_pb.GetEnvelopeReq{EnvelopeNo: createRedEnvelope.Data.EnvelopeNo, WaitSuccess: true})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "call wallet", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if envelopeResp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "get envelope", "error": envelopeResp.CommonResponse}).Send()
		return tcontext.RespToError(envelopeResp.CommonResponse)
	}

	if envelopeResp.Data.Status != pconst.EnvelopStatusRechargeSuccess {
		log.Error().Fields(map[string]interface{}{"action": fmt.Sprintf("create envelope failed, error:%s", err)}).Send()
		return he.NewBusinessError(0, text.EnvelopeCreateFailed, nil)
	}

	log.Debug().Fields(map[string]interface{}{"action": "create envelope success", "envelopeResp": envelopeResp})

	openButton := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text.OpenEnvelope, fmt.Sprintf("%s/%d", cmd.CmdOpenEnvelope, createRedEnvelope.Data.Id))},
	)
	ctx.TryDeleteMessage(pendingMsg)
	if _, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.CreateEnvelopeSuccess, createRedEnvelope.Data.Id, fmt.Sprintf("%s%s", pconst.GetExplore(payload.ChainType, pconst.ExploreTypeTx), createRedEnvelope.Data.TxHash)), nil, true, false); herr != nil {
		return herr
	}

	if replyMsg, herr := ctx.Send(channelId, fmt.Sprintf(text.ShareEnvelopeSuccess, ctx.GetNickNameMDV2(), createRedEnvelope.Data.Id, mdparse.ParseV2(payload.AssetSymbol), mdparse.ParseV2(payload.Amount)), &openButton, true, false); herr != nil {
		return herr
	} else {
		err = tstore.PBSaveString(fmt.Sprintf("%s%d", pconst.EnvelopeStorePrefix, createRedEnvelope.Data.Id), pconst.EnvelopeStorePath, strconv.FormatInt(int64(replyMsg.MessageID), 10))
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "TStore save envelope message error", "error": err.Error()}).Send()
		}
	}

	return nil
}
