package presetnode

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	pconst2 "github.com/tristan-club/wizard/pconst"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/log"
	"strconv"
	"strings"
)

var EnterAssetNode = chain.NewNode(askForAsset, prechecker.MustBeCallback, enterAsset)

type AssetConfigParam struct {
	AssetType    uint32 `json:"asset_type"`
	CheckBalance bool   `json:"check_balance"`
	SaveBalance  bool   `json:"save_balance"`
}

type AssetPayload struct {
	ChainType    uint32 `json:"chain_type"`
	ChainId      uint64 `json:"chain_id"`
	AssetAddress string `json:"asset_address"`
	AssetSymbol  string `json:"asset_symbol"`
	Decimals     uint32 `json:"decimals"`
}

func askForAsset(ctx *tcontext.Context, node *chain.Node) error {
	chainType, herr := userstate.MustUInt64(ctx.OpenId(), "chain_type")
	if herr != nil {
		return herr
	}
	var assetType uint32
	var param = &AssetConfigParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
		assetType = param.AssetType
	}
	assetListResp, err := ctx.CM.AssetList(ctx.Context, &controller_pb.AssetListReq{
		ChainType:    uint32(chainType),
		ChainId:      pconst2.GetChainId(uint32(chainType)),
		Address:      ctx.Requester.RequesterDefaultAddress,
		TokenType:    assetType,
		CheckBalance: param.CheckBalance,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if assetListResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(assetListResp.CommonResponse)
	} else if assetListResp.Data.Count == 0 {
		return he.NewBusinessError(0, text.NoAssetToOperation, nil)
	}

	var buttons []tgbotapi.InlineKeyboardButton
	for _, v := range assetListResp.Data.List {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(v.Symbol, fmt.Sprintf("%d/%d/%d/%s/%s", v.ChainType, v.ChainId, v.Decimals, v.Symbol, v.ContrAddr)))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
	thisMsg, herr := ctx.Send(ctx.U.SentFrom().ID, text.SelectAsset, &keyboard, false, false)
	if herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), thisMsg)
	}
	ctx.SetDeadlineMsg(ctx.U.SentFrom().ID, thisMsg.MessageID, pconst2.COMMON_KEYBOARD_DEADLINE)
	return nil
}

func enterAsset(ctx *tcontext.Context, node *chain.Node) error {
	assetPayload, herr := getAssetPayload(ctx.U.CallbackData())
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "get asset payload", "error": herr.Error()}).Send()
		return herr
	}

	var param = &AssetConfigParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}

	saveParam := map[string]interface{}{
		"asset":        assetPayload.AssetAddress,
		"asset_symbol": assetPayload.AssetSymbol,
	}

	if param.SaveBalance {
		assetResp, err := ctx.CM.GetAsset(ctx.Context, &controller_pb.GetAssetReq{
			ChainType:       assetPayload.ChainType,
			ChainId:         assetPayload.ChainId,
			Address:         ctx.Requester.RequesterDefaultAddress,
			ContractAddress: assetPayload.AssetAddress,
			TokenType:       0,
			ForceBalance:    true,
		})
		if err != nil {
			return he.NewServerError(he.CodeWalletRequestError, "", err)
		} else if assetResp.CommonResponse.Code != he.Success {
			return tcontext.RespToError(assetResp.CommonResponse)
		}
		saveParam["amount"] = assetResp.Data.Balance
		saveParam["amount_cut_decimal"] = assetResp.Data.BalanceCutDecimal
	}

	if herr := ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, ctx.U.CallbackQuery.Message.MessageID, fmt.Sprintf(text.ChosenAsset, assetPayload.AssetSymbol), nil, false, false); herr != nil {
		return herr
	}

	userstate.BatchSaveParam(ctx.OpenId(), saveParam)
	ctx.RemoveDeadlineMsg(ctx.U.CallbackQuery.Message.MessageID)
	return nil
}

func getAssetPayload(input string) (resp *AssetPayload, herr he.Error) {
	p := strings.Split(input, "/")
	if len(p) != 5 {
		return nil, he.NewServerError(he.CodeAssetParamInvalid, "", fmt.Errorf("invalid asset payload %s", input))

	}
	chainType, err := strconv.ParseUint(p[0], 10, 32)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "invalid payload", "error": err.Error(), "input": input}).Send()
		return nil, he.NewServerError(he.CodeInvalidPayload, "", err)
	}
	chainId, err := strconv.ParseUint(p[1], 10, 64)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "invalid payload", "error": err.Error(), "input": input}).Send()
		return nil, he.NewServerError(he.CodeInvalidPayload, "", err)
	}
	decimals, err := strconv.ParseUint(p[2], 10, 32)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "invalid payload", "error": err.Error(), "input": input}).Send()
		return nil, he.NewServerError(he.CodeInvalidPayload, "", err)
	}
	resp = &AssetPayload{
		ChainType:    uint32(chainType),
		ChainId:      chainId,
		Decimals:     uint32(decimals),
		AssetSymbol:  p[3],
		AssetAddress: p[4],
	}
	return resp, nil
}
