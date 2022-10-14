package cmd_balance

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/flow/presetnode"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/discordhandler/parser"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/pconst"
)

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options: []*discordgo.ApplicationCommandOption{
			presetnode.GetChainOption(),
		},
		Version: "1",
	},
	Handler: balanceSendHandler,
}

func balanceSendHandler(ctx *dcontext.Context) error {

	chainType, err := parser.OptionGetInt(ctx.IC.Interaction)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "invalid payload", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeInvalidPayload, "", err)
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
			content += fmt.Sprintf("%s\n%s\n\n", v.Symbol, v.BalanceCutDecimal)
		} else {
			content += fmt.Sprintf("%s(%s)\n%s\n\n", v.Symbol, v.ContrAddr, v.BalanceCutDecimal)
		}
	}

	if _, err = ctx.FollowUpReply(content); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}
	return nil
}
