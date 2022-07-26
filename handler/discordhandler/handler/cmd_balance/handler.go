package cmd_balance

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/bot-wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/flow/presetnode"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/handler"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/parser"
	"github.com/tristan-club/bot-wizard/handler/text"
	"github.com/tristan-club/bot-wizard/pconst"
	he "github.com/tristan-club/bot-wizard/pkg/error"
	"github.com/tristan-club/bot-wizard/pkg/log"
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

	if err := ctx.Reply(text.OperationProcessing); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg error", "error": err.Error()}).Send()
		return err
	}

	chainType, err := parser.OptionGetInt(ctx.IC.Interaction)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "invalid payload", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeInvalidPayload, "", err)
	}

	assetListResp, err := ctx.CM.AssetList(ctx.Context, &controller_pb.AssetListReq{
		ChainType:    uint32(chainType),
		ChainId:      uint64(pconst.GetChainId(uint32(chainType))),
		Address:      ctx.Requester.RequesterDefaultAddress,
		CheckBalance: true,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
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

	if _, err = ctx.EditReply(content); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}
	return nil
}
