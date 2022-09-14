package cmd_transfer

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/chain_info"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/flow/presetnode"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/discordhandler/parser"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/mdparse"
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

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options: []*discordgo.ApplicationCommandOption{
			presetnode.GetChainOption(),
			presetnode.GetAddressOption(&presetnode.OptionAddressPayload{Name: "to", Required: true}),
			presetnode.GetAmountOption(),
			presetnode.GetPinCodeOption("", ""),
		},
		Version: "1",
	},
	Handler: transferSendHandler,
}

func transferSendHandler(ctx *dcontext.Context) error {

	//if _, err := ctx.FollowUpReply(text.OperationProcessing); err != nil {
	//	log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
	//	return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	//}

	var payload = &TransferPayload{}

	err := parser.ParseOption(ctx.IC.Interaction, payload)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "parse param", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeInvalidPayload, "", err)
	}

	tokenType := pconst.TokenTypeInternal
	if payload.Asset != "" && len(payload.Asset) >= 40 {
		tokenType = pconst.TokenTypeErc20
	}

	req := &controller_pb.TransferReq{
		ChainType:       payload.ChainType,
		ChainId:         pconst.GetChainId(payload.ChainType),
		FromId:          ctx.Requester.RequesterUserNo,
		From:            ctx.Requester.RequesterDefaultAddress,
		To:              payload.To,
		ContractAddress: payload.Asset,
		TokenType:       uint32(tokenType),
		Nonce:           0,
		Value:           payload.Amount,
		GasLimit:        0,
		GasPrice:        0,
		PinCode:         payload.PinCode,
		CheckBalance:    true,
		IsWait:          false,
	}

	transferResp, err := ctx.CM.Transfer(ctx.Context, req)
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if transferResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(transferResp.CommonResponse)
	}

	if _, err = ctx.FollowUpReply(fmt.Sprintf(text.TransactionProcessing, pconst.GetExplore(payload.ChainType, transferResp.Data.TxHash, chain_info.ExplorerTargetTransaction))); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	getDataResp, err := ctx.CM.GetTx(context.Background(), &controller_pb.GetTxReq{TxHash: transferResp.Data.TxHash})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if getDataResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(getDataResp.CommonResponse)
	}

	content := fmt.Sprintf(text.TransferSuccess, payload.To, mdparse.ParseV2(payload.AssetSymbol),
		mdparse.ParseV2(payload.Amount), pconst.GetExplore(payload.ChainType, transferResp.Data.TxHash, chain_info.ExplorerTargetTransaction))

	if err := ctx.DM(content); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	return nil
}
