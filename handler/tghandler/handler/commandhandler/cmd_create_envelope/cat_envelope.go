package cmd_create_envelope

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tristan-club/kit/chain_info"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/inline_keybord"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pconst"
)

var CATEnvelopeHandler *chain.ChainHandler

var selectChainNode = chain.NewNode(askForChain, prechecker.MustBeCallback, presetnode.EnterChain)

func askForChain(ctx *tcontext.Context, node *chain.Node) error {

	param, herr := userstate.GetParam(ctx.OpenId(), "chain_type_list")
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "get user state error", "error": herr.Error()}).Send()
		return herr
	}

	b, _ := json.Marshal(param)

	var intList []uint32
	if err := json.Unmarshal(b, &intList); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "get chain type list error", "error": err.Error(), "param": param}).Send()
		return he.NewServerError(he.ServerError, "", err)
	}
	thisMsg, herr := ctx.Send(ctx.U.SentFrom().ID, text.SelectChain, inline_keybord.GenerateKeyBoardByChainTypeList(intList), false, false)
	if herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), thisMsg)
	}
	ctx.SetDeadlineMsg(ctx.U.SentFrom().ID, thisMsg.MessageID, pconst.COMMON_KEYBOARD_DEADLINE)
	return nil
}

func catChecker(ctx *tcontext.Context) error {

	assetListResp, err := ctx.CM.AssetList(context.Background(), &controller_pb.AssetListReq{
		AppId:      ctx.Requester.RequesterAppId,
		PresetType: controller_pb.AssetPresetType_PRESET_TYPE_CAT,
		TokenType:  pconst.TokenTypeERC721,
	})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "request controller svc error", "error": err.Error(), "ctx": ctx}).Send()
		return he.NewServerError(he.ServerError, "", err)
	} else if assetListResp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "get cat error", "error": assetListResp, "ctx": ctx}).Send()
		return he.NewServerError(he.ServerError, assetListResp.CommonResponse.Message, fmt.Errorf(assetListResp.CommonResponse.Inner))
	} else if assetListResp.Data.Count == 0 {
		log.Info().Fields(map[string]interface{}{"action": "current app not configure cat", "ctx": ctx}).Send()
		return he.NewBusinessError(pconst.CodeCATNotConfigure, "", nil)
	}

	var chainIdList []uint32
	for _, v := range assetListResp.Data.List {
		chainIdList = append(chainIdList, chain_info.GetNetByChainId(v.ChainId).ChainType)
	}

	userstate.SetParam(ctx.OpenId(), "envelope_option", uint32(controller_pb.ENVELOPE_OPTION_HAS_CAT))
	userstate.SetParam(ctx.OpenId(), "chain_type_list", chainIdList)

	return nil
}

func init() {
	enterEnvelopeTypeNode = new(chain.Node)
	*enterEnvelopeTypeNode = *presetnode.EnterTypeNode

	enterEnvelopeQuantityNode := chain.NewNode(presetnode.AskForQuantity, prechecker.MustBeMessage, presetnode.EnterQuantity)

	CATEnvelopeHandler = chain.NewChainHandler(cmd.CmdCreateEnvelope, createEnvelopeSendHandler).
		AddPreHandler(prehandler.OnlyPublic).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPreHandler(catChecker).
		AddPresetNode(selectChainNode, nil).
		AddPresetNode(presetnode.EnterAssetNode, nil)

	if config.EnableTaskEnvelope() {
		CATEnvelopeHandler.AddPresetNode(enterEnvelopeTypeNode, &presetnode.EnterTypeParam{
			ChoiceText:  envelopeTypeText,
			ChoiceValue: envelopeTypeValue,
			Content:     text.SelectEnvelopeType,
			ParamKey:    "envelope_type",
		})
	}

	CATEnvelopeHandler.AddPresetNode(enterEnvelopeTypeNode, &presetnode.EnterTypeParam{
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
