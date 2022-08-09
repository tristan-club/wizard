package bridge_node

import (
	"fmt"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/bignum"
	he "github.com/tristan-club/wizard/pkg/error"
	"math/big"
	"strings"
)

var EnterBridgeAmountNode = chain.NewNode(askForBridgeAmount, nil, enterBridgeAmount)
var fromChainType = uint32(chain_info.ChainTypeBsc)
var fromChainId = pconst.GetChainId(fromChainType)
var toChainType = uint32(chain_info.ChainTypeMetis)
var toChainId = pconst.GetChainId(toChainType)

var bridgeAmountMin *big.Int
var bridgeAmountMax *big.Int
var bridgeAmountMinStr = "0.001"
var bridgeAmountMinStrLabel = "0\\.001"
var bridgeAmountMaxStr = "2"
var bridgeAssetSymbol = "METIS"
var bridgeGas = "0.001"
var bridgeGasBig *big.Int
var bridgeFee = "0.01"
var bridgeFeeBig *big.Int
var bridgeMinAvaliableAmount = "0.01"
var bridgeMinAvaliableAmountBig *big.Int
var tokenAddress = pconst.GetSwapAssetAddress(bridgeAssetSymbol)

func init() {
	bridgeAmountMin, _ = bignum.HandleAddDecimal(bridgeAmountMinStr, 18)
	bridgeAmountMax, _ = bignum.HandleAddDecimal(bridgeAmountMaxStr, 18)
	bridgeGasBig, _ = bignum.HandleAddDecimal(bridgeGas, 18)
	bridgeFeeBig, _ = bignum.HandleAddDecimal(bridgeFee, 18)
	bridgeMinAvaliableAmountBig, _ = bignum.HandleAddDecimal(bridgeMinAvaliableAmount, 18)
}

func askForBridgeAmount(ctx *tcontext.Context, node *chain.Node) error {

	assetResp, err := ctx.CM.GetAsset(ctx.Context, &controller_pb.GetAssetReq{
		ChainType:       fromChainType,
		ChainId:         fromChainId,
		Address:         ctx.Requester.RequesterDefaultAddress,
		ContractAddress: tokenAddress,
		ForceBalance:    true,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if assetResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(assetResp.CommonResponse)
	}

	balanceBig, ok := new(big.Int).SetString(assetResp.Data.Balance, 10)
	if !ok {
		log.Error().Fields(map[string]interface{}{"action": "invalid balance", "error": assetResp}).Send()
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	}

	//content :=

	if balanceBig.Cmp(bridgeAmountMin) < 0 {
		content := fmt.Sprintf(text.InSufficientBalance, "METIS", "BNB", ctx.Requester.RequesterDefaultAddress)
		if _, herr := ctx.Send(ctx.U.SentFrom().ID, content, nil, true, false); herr != nil {
			return herr
		}
		return nil
	}

	//tokenAddressLabel := "\\(`" + tokenAddress + "`\\)"

	balance := assetResp.Data.BalanceCutDecimal

	bankBalanceResp, err := ctx.CM.BankBalance(ctx.Context, &controller_pb.GetBankBalanceReq{
		ChainType: toChainType,
		ChainId:   toChainId,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if bankBalanceResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(bankBalanceResp.CommonResponse)
	}

	bankBalance, ok := new(big.Int).SetString(bankBalanceResp.Data.Balance, 10)
	if !ok {
		log.Error().Fields(map[string]interface{}{"action": "invalid bank balance", "balance": bankBalance}).Send()
		return he.NewServerError(he.CodeBankLackBalance, "", fmt.Errorf("invalid bank balance %s", bankBalanceResp.Data.Balance))
	}

	if bankBalance.Cmp(bridgeMinAvaliableAmountBig) < 0 {
		log.Warn().Fields(map[string]interface{}{"action": "bank balance low", "balance": bankBalance})
		return he.NewBusinessError(he.CodeBankLackBalance, "", nil)
	}

	content := fmt.Sprintf(text.EnterBridgeAmount, balance, bridgeAmountMinStr, bridgeAmountMaxStr)
	if strings.Contains(content, ".") {
		content = strings.ReplaceAll(content, ".", "\\.")
	}

	//ikm, deadlineTime := inline_keybord.NewMaxAmountKeyboard()

	if msg, herr := ctx.Send(ctx.U.SentFrom().ID, content, nil, true, false); herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), msg)
		//inline_keybord.DeleteDeadKeyboard(ctx, deadlineTime, msg)
		return nil
	}

}

func enterBridgeAmount(ctx *tcontext.Context, node *chain.Node) error {
	swapAsset := pconst.GetBridgeAsset(toChainType, bridgeAssetSymbol)

	amountInput := ctx.U.Message.Text
	amountInputBig, ok := bignum.HandleAddDecimal(amountInput, 18)
	if !ok {
		return he.NewBusinessError(he.CodeAmountParamInvalid, "", nil)
	}

	if amountInputBig.Cmp(bridgeAmountMin) < 0 || amountInputBig.Cmp(bridgeAmountMax) > 0 {
		return he.NewBusinessError(he.CodeBridgeAmountMustInScope, "", nil)
	}

	assetResp, err := ctx.CM.GetAsset(ctx.Context, &controller_pb.GetAssetReq{
		ChainType:       fromChainType,
		ChainId:         fromChainId,
		Address:         ctx.Requester.RequesterDefaultAddress,
		ContractAddress: swapAsset.AssetAddress,
		ForceBalance:    true,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if assetResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(assetResp.CommonResponse)
	}

	balanceBig, ok := new(big.Int).SetString(assetResp.Data.Balance, 10)
	if !ok {
		return he.NewServerError(he.CodeWalletRequestError, "", fmt.Errorf("invalid balance %s", assetResp.Data.BalanceCutDecimal))
	}

	if amountInputBig.Cmp(balanceBig) > 0 {
		return he.NewBusinessError(he.CodeInsufficientBalance, fmt.Sprintf("Insufficient amount input\nYour balance %s\nYour input %s", assetResp.Data.BalanceCutDecimal, amountInput), nil)
	}

	total := new(big.Int).Add(amountInputBig, bridgeFeeBig)
	if total.Cmp(balanceBig) > 0 {
		totalStr := bignum.HandleDecimal(total, 18)
		t2, _ := bignum.HandleAddDecimal(totalStr, 18)
		t3, _ := bignum.HandleAddDecimal(assetResp.Data.BalanceCutDecimal, 18)
		if t3.Cmp(t2) != 0 {
			return he.NewBusinessError(he.CodeInsufficientBalance, fmt.Sprintf("Insufficient amount input\nYour balance %s\nTotal cost %s\nYou need %s for bridge value and %s for bridge fee",
				assetResp.Data.BalanceCutDecimal, totalStr, amountInput, bridgeFee), nil)
		}

	}

	//totalAmount := new(big.Int).Add(amountInputBig, bridgeFeeBig)
	//totalAmount = new(big.Int).Add(totalAmount, bridgeGasBig)
	//log.Debug().Fields(map[string]interface{}{"action": "total amount", "amount": totalAmount.String()}).Send()

	bankBalanceResp, err := ctx.CM.BankBalance(ctx.Context, &controller_pb.GetBankBalanceReq{
		ChainType: toChainType,
		ChainId:   toChainId,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if assetResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(assetResp.CommonResponse)
	}

	if bankBalanceResp.Data.Balance == "" || bankBalanceResp.Data.Balance == "0" {
		return he.NewBusinessError(he.CodeNoBankBalance, "", nil)
	} else {
		bankBalanceBig, ok := new(big.Int).SetString(bankBalanceResp.Data.Balance, 10)
		if !ok {
			return he.NewServerError(he.CodeInsufficientBalance, "", fmt.Errorf("get bank balance error"))
		}
		if bankBalanceBig.Cmp(amountInputBig) < 0 {
			return he.NewBusinessError(he.CodeBankLackBalance, "", nil)
		}
	}

	//gasAmountBig, ok := bignum.HandleAddDecimal(pconst.BridgeGasAmount, 18)
	//if !ok {
	//	return he.NewServerError(fmt.Errorf("invalid bridge fee %s", pconst.BridgeGasAmount))
	//}
	//toAmount := new(big.Int).Sub(amountInputBig, gasAmountBig)
	//toAmountStr := bignum.HandleDecimal(toAmount, 18)
	confirmContent := fmt.Sprintf(text.BridgeConfirmOrder, amountInput, bridgeGas, bridgeFee)
	userstate.SetParam(ctx.OpenId(), "amount", amountInput)
	userstate.SetParam(ctx.OpenId(), "bridge_content", confirmContent)
	return nil
}
