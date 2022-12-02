package cmd_envelope

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/mdparse"
	"github.com/tristan-club/kit/tstore"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"strconv"
	"strings"
)

const (
	EnvelopeNumMin = 1
	EnvelopeNumMax = 20

	AmountMin = "0.0001"
	AmountMax = "10000000000"
)

var envelopeRewardTypeText = []string{"Average Amount", "Random Amount"}
var envelopeRewardTypeValue = []int64{1, 2}

var envelopeTypeText = []string{"Ordinary Red Envelope", "Task Red Envelopee"}
var envelopeTypeValue = []int64{pconst.EnvelopeTypeOrdinary, pconst.EnvelopeTypeTask}

type CreateEnvelopePayload struct {
	UserNo             string   `json:"user_no"`
	From               string   `json:"from"`
	ChainType          uint32   `json:"chain_type"`
	Asset              string   `json:"asset"`
	AssetSymbol        string   `json:"asset_symbol"`
	EnvelopeRewardType uint32   `json:"envelope_reward_type"`
	EnvelopeType       uint32   `json:"envelope_type"`
	ChannelId          string   `json:"channel_id"`
	ChannelUsername    string   `json:"channel_username"`
	Quantity           uint64   `json:"quantity"`
	Amount             string   `json:"amount"`
	PinCode            string   `json:"pin_code"`
	EnvelopeOption     uint32   `json:"envelope_option"`
	ChainTypeList      []uint32 `json:"chain_type_list"`
}

var CreateOrdinaryEnvelopeHandler *chain.ChainHandler

var enterEnvelopeTypeNode *chain.Node

func init() {
	enterEnvelopeTypeNode = new(chain.Node)
	*enterEnvelopeTypeNode = *presetnode.EnterTypeNode

	enterEnvelopeQuantityNode := chain.NewNode(presetnode.AskForQuantity, prechecker.MustBeMessage, presetnode.EnterQuantity)

	CreateOrdinaryEnvelopeHandler = chain.NewChainHandler(cmd.CmdCreateEnvelope, createEnvelopeSendHandler).
		AddPreHandler(prehandler.OnlyPublic).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.SelectChainNode, nil).
		AddPresetNode(presetnode.EnterAssetNode, nil)

	if config.EnableTaskEnvelope() {
		CreateOrdinaryEnvelopeHandler.AddPresetNode(enterEnvelopeTypeNode, &presetnode.EnterTypeParam{
			ChoiceText:  envelopeTypeText,
			ChoiceValue: envelopeTypeValue,
			Content:     text.SelectEnvelopeType,
			ParamKey:    "envelope_type",
		})
	}

	CreateOrdinaryEnvelopeHandler.AddPresetNode(enterEnvelopeTypeNode, &presetnode.EnterTypeParam{
		ChoiceText:  envelopeRewardTypeText,
		ChoiceValue: envelopeRewardTypeValue,
		Content:     text.SelectEnvelopeRewardType,
		ParamKey:    "envelope_reward_type",
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

	msg, herr := ctx.Send(ctx.U.SentFrom().ID, text.OperationProcessing, nil, false, false)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg error", "error": herr}).Send()
		return herr
	}

	defer ctx.TryDeleteMessage(msg)

	tokenType := pconst.TokenTypeInternal
	if payload.Asset != "" && payload.Asset != "INTERNAL" && strings.HasPrefix(payload.Asset, "0x") {
		tokenType = pconst.TokenTypeErc20
	}

	createEnvelopeReq := &controller_pb.AddEnvelopeReq{
		FromId:             payload.UserNo,
		ChainType:          payload.ChainType,
		ChannelId:          payload.ChannelId,
		ChainId:            pconst.GetChainId(payload.ChainType),
		TokenType:          uint32(tokenType),
		Address:            payload.From,
		ContractAddress:    payload.Asset,
		Amount:             payload.Amount,
		Quantity:           payload.Quantity,
		EnvelopeType:       payload.EnvelopeRewardType,
		EnvelopeRewardType: payload.EnvelopeRewardType,
		Blessing:           "",
		PinCode:            payload.PinCode,
		IsWait:             false,
		EnvelopeOption:     controller_pb.ENVELOPE_OPTION(payload.EnvelopeOption),
	}

	createRedEnvelope, err := ctx.CM.AddEnvelope(ctx.Context, createEnvelopeReq)
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if createRedEnvelope.CommonResponse.Code != he.Success {
		return tcontext.RespToError(createRedEnvelope.CommonResponse)
	}

	pendingMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.EnvelopePreparing, mdparse.ParseV2(pconst.GetExplore(payload.ChainType, createRedEnvelope.Data.AccountAddress, chain_info.ExplorerTargetAddress))), nil, true, false)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "send pending tx error", "error": herr.Error()}).Send()
		return herr
	}

	defer ctx.TryDeleteMessage(pendingMsg)

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

	cid := customid.NewCustomId(pconst.CustomIdOpenEnvelope, createRedEnvelope.Data.EnvelopeNo, int32(payload.EnvelopeOption))
	openButton := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text.OpenEnvelope, cid.String())},
	)

	if _, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.CreateEnvelopeSuccess, createRedEnvelope.Data.EnvelopeNo, mdparse.ParseV2(pconst.GetExplore(payload.ChainType, createRedEnvelope.Data.TxHash, chain_info.ExplorerTargetTransaction))), nil, true, false); herr != nil {
		return herr
	}

	title := text.EnvelopeTitleOrdinary
	if payload.EnvelopeOption == uint32(controller_pb.ENVELOPE_OPTION_HAS_CAT) {
		title = text.EnvelopeTitleCAT
	}
	title = fmt.Sprintf(title, envelopeResp.Data.EnvelopeNo, ctx.GetNickNameMDV2())

	shareEnvelopeContent := fmt.Sprintf(text.EnvelopeDetail, mdparse.ParseV2(payload.Amount), mdparse.ParseV2(payload.AssetSymbol), 0, payload.Quantity)

	shareEnvelopeContent = title + "\n\n" + shareEnvelopeContent

	if replyMsg, herr := ctx.Send(channelId, shareEnvelopeContent, &openButton, true, false); herr != nil {
		return herr
	} else {
		err = tstore.PBSaveStr(fmt.Sprintf("%s%s", pconst.EnvelopeStorePrefix, createRedEnvelope.Data.EnvelopeNo), pconst.EnvelopeStorePathMsgId, strconv.FormatInt(int64(replyMsg.MessageID), 10))
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "TStore save envelope message error", "error": err.Error()}).Send()
		}

		err = tstore.PBSaveStr(fmt.Sprintf("%s%s", pconst.EnvelopeStorePrefix, createRedEnvelope.Data.EnvelopeNo), pconst.EnvelopeStorePathChannelId, fmt.Sprintf("%s/%s", payload.ChannelId, payload.ChannelUsername))
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "TStore save envelope channel id error", "error": err.Error()}).Send()
		}

	}

	return nil
}
