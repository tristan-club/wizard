package cmd_transfer

import (
	"context"
	"fmt"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	he "github.com/tristan-club/wizard/pkg/error"
	"strings"
)

type TransferPayload struct {
	UserNo      string `json:"user_no"`
	From        string `json:"from"`
	To          string `json:"to"`
	AssetSymbol string `json:"asset_symbol"`
	ChainType   uint32 `json:"chain_type"`
	Asset       string `json:"asset"`
	Amount      string `json:"amount"`
	PinCode     string `json:"pin_code"`
}

var enterTransferAmountHandler *chain.Node

var Handler *chain.ChainHandler

func init() {
	enterTransferAmountHandler = new(chain.Node)
	*enterTransferAmountHandler = *presetnode.EnterAmountNode
	Handler = chain.NewChainHandler(cmd.CmdTransfer, transferSendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.SelectChainNode, nil).
		AddPresetNode(presetnode.EnterAssetNode, nil).
		AddPresetNode(presetnode.EnterAddressNode, nil).
		AddPresetNode(enterTransferAmountHandler, &presetnode.AmountParam{CheckBalance: false, WithMaxButton: true}).
		AddPresetNode(presetnode.EnterPinCodeHandler, nil)
}

func transferSendHandler(ctx *tcontext.Context) error {

	var payload = &TransferPayload{}

	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}
	tokenType := pconst.TokenTypeInternal
	if payload.Asset != "" && len(payload.Asset) >= 40 {
		tokenType = pconst.TokenTypeErc20
	}

	req := &controller_pb.TransferReq{
		ChainType:       payload.ChainType,
		ChainId:         pconst.GetChainId(payload.ChainType),
		FromId:          payload.UserNo,
		From:            payload.From,
		To:              payload.To,
		ContractAddress: payload.Asset,
		TokenType:       uint32(tokenType),
		Nonce:           0,
		Value:           "",
		GasLimit:        0,
		GasPrice:        0,
		PinCode:         payload.PinCode,
		CheckBalance:    true,
		IsWait:          false,
	}
	if payload.Amount == pconst.MaxAmount {
		req.MaxAmount = true
	} else {
		req.Value = payload.Amount
	}

	transferResp, err := ctx.CM.Transfer(ctx.Context, req)
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if transferResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(transferResp.CommonResponse)
	}

	thisMsg, herr := ctx.Send(ctx.U.SentFrom().ID, fmt.Sprintf(text.TransactionProcessing, fmt.Sprintf("%s%s", pconst.GetExplore(payload.ChainType, pconst.ExploreTypeTx), transferResp.Data.TxHash)), nil, true, false)
	if herr != nil {
		return herr
	}

	getDataResp, err := ctx.CM.GetTx(context.Background(), &controller_pb.GetTxReq{TxHash: transferResp.Data.TxHash})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if getDataResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(getDataResp.CommonResponse)
	}

	content := fmt.Sprintf(text.TransferSuccess, payload.To, payload.AssetSymbol, payload.Amount, fmt.Sprintf("%s%s", pconst.GetExplore(payload.ChainType, pconst.ExploreTypeTx), transferResp.Data.TxHash))
	content = strings.ReplaceAll(content, ".", "\\.")
	herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, content, nil, true, false)
	if herr != nil {
		return herr
	}
	return nil
}
