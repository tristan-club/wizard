package swap_node

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/bignum"
	he "github.com/tristan-club/wizard/pkg/error"
	"math/big"
	"strings"
)

const (
	Slider_Point_Max = 0.003
)

var EnterSwapAndBridgeAmountNode = chain.NewNode(askForSwapAmount, prechecker.MustBeMessage, enterSwapAmount)
var toChainType = uint32(chain_info.ChainTypeBsc)
var toChainId = pconst.GetChainId(toChainType)

func askForSwapAmount(ctx *tcontext.Context, _ *chain.Node) error {

	chainType := uint32(chain_info.ChainTypeBsc)
	chainId := pconst.GetChainId(chainType)

	swapAssetSymbol, herr := userstate.MustString(ctx.OpenId(), "asset_symbol")
	if herr != nil {
		return herr
	}
	contractAddress := pconst.GetSwapAssetAddress(swapAssetSymbol)

	assetResp, err := ctx.CM.GetAsset(ctx.Context, &controller_pb.GetAssetReq{
		ChainType:       chainType,
		ChainId:         chainId,
		Address:         ctx.Requester.RequesterDefaultAddress,
		ContractAddress: contractAddress,
		ForceBalance:    true,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if assetResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(assetResp.CommonResponse)
	}
	swapAsset := pconst.GetSwapAsset(chainType, swapAssetSymbol)

	content := generateEnterAmountContent(swapAssetSymbol, assetResp.Data.BalanceCutDecimal, swapAsset.SwapAmountMin, swapAsset.SwapAmountMax)
	content = strings.ReplaceAll(content, ".", "\\.")

	if _, herr = ctx.Send(ctx.U.SentFrom().ID, content, nil, true, false); herr != nil {
		return herr
	}
	return nil
}

func enterSwapAmount(ctx *tcontext.Context, node *chain.Node) error {
	chainType := uint32(chain_info.ChainTypeBsc)
	chainId := pconst.GetChainId(chainType)
	assetSymbol, herr := userstate.MustString(ctx.OpenId(), "asset_symbol")
	if herr != nil {
		return herr
	}
	swapAsset := pconst.GetSwapAsset(chainType, assetSymbol)

	amountInput := ctx.U.Message.Text
	amountInputBig, ok := bignum.HandleAddDecimal(amountInput, int(swapAsset.Decimals))
	if !ok {
		return he.NewBusinessError(he.CodeAmountParamInvalid, "", nil)
	}
	min, ok := bignum.HandleAddDecimal(swapAsset.SwapAmountMin, int(swapAsset.Decimals))
	if !ok {
		log.Error().Msgf("preset amount param min %s invalid, node id %s", swapAsset.SwapAmountMin, node.Id)
		return he.NewBusinessError(he.CodeAmountParamInvalid, "", nil)
	}
	max, ok := bignum.HandleAddDecimal(swapAsset.SwapAmountMax, int(swapAsset.Decimals))
	if !ok {
		log.Error().Msgf("preset amount param max %s invalid, node id %s", swapAsset.SwapAmountMax, node.Id)
		return he.NewBusinessError(he.CodeAmountParamInvalid, "", nil)
	}
	if amountInputBig.Cmp(min) < 0 || amountInputBig.Cmp(max) > 0 {
		return he.NewBusinessError(he.CodeBridgeAmountMustInScope, "", nil)
	}

	assetResp, err := ctx.CM.GetAsset(ctx.Context, &controller_pb.GetAssetReq{
		ChainType:       chainType,
		ChainId:         chainId,
		Address:         ctx.Requester.RequesterDefaultAddress,
		ContractAddress: swapAsset.AssetAddress,
		ForceBalance:    true,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if assetResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(assetResp.CommonResponse)
	}

	balance, ok := new(big.Int).SetString(assetResp.Data.Balance, 10)

	if !ok {
		return he.NewServerError(he.CodeWalletRequestError, "", fmt.Errorf("invalid balance %s", assetResp.Data.Balance))
	}

	if amountInputBig.Cmp(balance) > 0 {
		return he.NewBusinessError(he.CodeInsufficientBalance, "", nil)
	}
	var tokenType uint32
	if assetResp.GetData().TokenType != 0 {
		tokenType = assetResp.GetData().TokenType
	} else {
		tokenType = pconst.TokenTypeErc20
	}

	swapCalculateReq := &controller_pb.SwapCalculateReq{
		ChainType:          toChainType,
		ChainId:            toChainId,
		FromId:             ctx.OpenId(),
		From:               ctx.Requester.RequesterDefaultAddress,
		TokenType:          tokenType,
		OriginTokenAddress: swapAsset.AssetAddress,
		Value:              amountInput,
		TargetTokenAddress: pconst.TOKEN_BSC_METIS,
		ContractAddress:    pconst.PANCAKE_BSC,
		ChannelId:          ctx.Requester.RequesterChannelId,
	}

	if swapAsset.AssetAddress == "INTERNAL" || len(swapAsset.AssetAddress) == 0 {
		swapCalculateReq.TokenType = pconst.TokenTypeInternal
	}
	swapCalculateResp, err := ctx.CM.SwapCalculate(ctx.Context, swapCalculateReq)

	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if swapCalculateResp.Code != he.Success {
		return he.NewBusinessError(0, swapCalculateResp.Message, nil)
	}
	outDecimal, err := decimal.NewFromString(swapCalculateResp.CurrentOut)
	if err != nil {
		log.Error().Msgf("convert calculate out error:%s", err)
		return he.NewServerError(he.CodeWalletRequestError, "", fmt.Errorf("calculate error %s", err.Error()))
	}
	outMin := outDecimal.Mul(decimal.NewFromFloat(1 - Slider_Point_Max))

	content := generateOrderInfo(swapAsset.AssetSymbol, swapCalculateReq.Value, outDecimal, "0.0028")
	userstate.SetParam(ctx.OpenId(), "amount", amountInput)
	userstate.SetParam(ctx.OpenId(), "swap_content", "")
	userstate.SetParam(ctx.OpenId(), "swap_content", content)
	userstate.SetParam(ctx.OpenId(), "chain_type", toChainType)
	userstate.SetParam(ctx.OpenId(), "chain_id", toChainId)
	userstate.SetParam(ctx.OpenId(), "token_type", swapCalculateReq.TokenType)
	userstate.SetParam(ctx.OpenId(), "origin_token", swapCalculateReq.OriginTokenAddress)
	userstate.SetParam(ctx.OpenId(), "target_token", pconst.TOKEN_BSC_METIS)
	userstate.SetParam(ctx.OpenId(), "target_amount", outMin.String())

	return nil
}

func generateEnterAmountContent(token string, balance string, inputMin string, inputMax string) string {
	var content string
	content = fmt.Sprintf(text.SwapEnterAmount, token, balance, inputMin, inputMax)
	return content
}

func generateOrderInfo(originalAsset string, amountIn string, out decimal.Decimal, gas string) string {
	var content string
	outMin := out.Mul(decimal.NewFromFloat(1 - Slider_Point_Max))
	outMax := out.Mul(decimal.NewFromFloat(1 + Slider_Point_Max))

	content = fmt.Sprintf(text.SwapOrder, amountIn, originalAsset, gas, outMin.String()[0:8], outMax.String()[0:8])
	return content
}
