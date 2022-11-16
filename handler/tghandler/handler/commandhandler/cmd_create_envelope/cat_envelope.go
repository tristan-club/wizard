package cmd_create_envelope

import (
	"context"
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
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
)

var CATEnvelopeHandler *chain.ChainHandler

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
	userstate.SetParam(ctx.OpenId(), "cat_chain_list", chainIdList)

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
		AddPresetNode(presetnode.SelectChainNode, nil).
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
