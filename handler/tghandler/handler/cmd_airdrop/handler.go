package cmd_airdrop

import (
	"context"
	"fmt"
	"github.com/tristan-club/kit/chain_info"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/mdparse"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"strconv"
)

type AirdropPayload struct {
	UserNo      string `json:"user_no"`
	From        string `json:"from"`
	ChainType   uint32 `json:"chain_type"`
	Asset       string `json:"asset"`
	AssetSymbol string `json:"asset_symbol"`
	Amount      string `json:"amount"`
	PinCode     string `json:"pin_code"`
	ChannelId   string `json:"channel_id"`
}

var Handler *chain.ChainHandler

func init() {

	Handler = chain.NewChainHandler(cmd.CmdAirdrop, airdropSendHandler).
		AddPreHandler(prehandler.OnlyPublic).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.SelectChainNode, nil).
		AddPresetNode(presetnode.EnterAssetNode, presetnode.AssetConfigParam{AssetType: pconst.TokenTypeErc20}).
		AddPresetNode(presetnode.EnterAmountNode, presetnode.AmountParam{
			CheckBalance: true,
			Content:      text.EnterAmount,
			ParamKey:     "amount",
		}).
		AddPresetNode(presetnode.EnterPinCodeHandler, nil)
}

func airdropSendHandler(ctx *tcontext.Context) error {

	var payload = &AirdropPayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	channelId, err := strconv.ParseInt(payload.ChannelId, 10, 64)
	if err != nil {
		return he.NewServerError(pconst.CodeInvalidPayload, "", err)
	}

	req := &controller_pb.AirdropReq{
		ChainType:       payload.ChainType,
		ChainId:         pconst.GetChainId(payload.ChainType),
		Address:         payload.From,
		FromId:          payload.UserNo,
		PinCode:         payload.PinCode,
		TokenType:       pconst.TokenTypeErc20,
		ContractAddress: payload.Asset,
		Amount:          payload.Amount,
		ChannelId:       payload.ChannelId,
	}

	transactionResp, err := ctx.CM.Airdrop(ctx.Context, req)
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if transactionResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(transactionResp.CommonResponse)
	}

	thisMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.TransactionProcessing, mdparse.ParseV2(pconst.GetExplore(payload.ChainType, transactionResp.Data.TxHash, chain_info.ExplorerTargetTransaction))), nil, true, false)
	if herr != nil {
		return herr
	}

	getDataResp, err := ctx.CM.GetTx(context.Background(), &controller_pb.GetTxReq{TxHash: transactionResp.Data.TxHash, IsWait: true})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if getDataResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(getDataResp.CommonResponse)
	}

	if herr = ctx.EditMessageAndKeyboard(ctx.U.FromChat().ID, thisMsg.MessageID, fmt.Sprintf(text.AirdropSuccess, mdparse.ParseV2(pconst.GetExplore(payload.ChainType, getDataResp.Data.TxHash, chain_info.ExplorerTargetTransaction))), nil, true, false); herr != nil {
		return herr
	}

	receiverList := ""
	for k, v := range transactionResp.Data.ReceiverList {
		receiverList += fmt.Sprintf("\\- [@%s](tg://user?id=%s)\n", mdparse.ParseV2(v), transactionResp.Data.OpenIdList[k])
	}
	content := fmt.Sprintf(text.AirdropSuccessInGroup, ctx.GetNickNameMDV2(), mdparse.ParseV2(payload.Amount), mdparse.ParseV2(payload.AssetSymbol),
		mdparse.ParseV2(transactionResp.Data.Amount), mdparse.ParseV2(payload.AssetSymbol), receiverList, mdparse.ParseV2(pconst.GetExplore(payload.ChainType, getDataResp.Data.TxHash, chain_info.ExplorerTargetTransaction)))
	if _, herr := ctx.Send(channelId, content, nil, true, true); herr != nil {
		return herr
	}

	return nil
}
