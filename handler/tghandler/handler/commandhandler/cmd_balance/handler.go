package cmd_balance

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
)

var Handler = chain.NewChainHandler(cmd.CmdBalance, balanceSendHandler).
	AddCmdParser(func(u *tgbotapi.Update) string {
		if u.Message != nil && u.Message.Text == text.KBBalance {
			return cmd.CmdBalance
		}
		return ""
	}).
	AddPreHandler(prehandler.ForwardPrivate).AddPresetNode(presetnode.SelectChainNode, nil)

func balanceSendHandler(ctx *tcontext.Context) error {

	chainType, herr := userstate.MustUInt64(ctx.OpenId(), "chain_type")
	if herr != nil {
		return herr
	}

	assetListResp, err := ctx.CM.AssetList(ctx.Context, &controller_pb.AssetListReq{
		ChainType:    uint32(chainType),
		ChainId:      uint64(pconst.GetChainId(uint32(chainType))),
		Address:      ctx.Requester.RequesterDefaultAddress,
		CheckBalance: true,
		TokenType:    pconst.TokenTypeCoinOrERC20,
	})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if assetListResp.CommonResponse.Code != he.Success {
		return he.NewServerError(int(assetListResp.CommonResponse.Code), "", fmt.Errorf(assetListResp.CommonResponse.Message))
	}
	content := text.BalanceSuccess
	content += "\n"
	for _, v := range assetListResp.Data.List {
		if v.TokenType == pconst.TokenTypeInternal {
			if v.TotalPrice != "" {
				content += fmt.Sprintf("%s\n%s( ~ $%s)\n\n", v.Symbol, v.BalanceCutDecimal, v.TotalPrice)
			} else {
				content += fmt.Sprintf("%s\n%s\n\n", v.Symbol, v.BalanceCutDecimal)
			}

		} else {
			content += fmt.Sprintf("%s(%s)\n%s\n\n", v.Symbol, v.ContrAddr, v.BalanceCutDecimal)
		}

	}

	_, herr = ctx.Send(ctx.U.FromChat().ID, content, nil, false, false)
	if herr != nil {
		return herr
	}
	return nil
}
