package cmd_add_token

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/flow/presetnode"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/discordhandler/parser"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/log"
)

type AddTokenPayload struct {
	ChainType       uint32 `json:"chain_type"`
	TokenType       uint32 `json:"token_type"`
	ContractAddress string `json:"contract_address"`
}

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options: []*discordgo.ApplicationCommandOption{
			presetnode.GetChainOption(),
			presetnode.GetAddressOption(&presetnode.OptionAddressPayload{Name: "contract_address"}),
		},
		Version: "1",
	},
	Handler: addTokenSendHandler,
}

func addTokenSendHandler(ctx *dcontext.Context) error {

	var payload = &AddTokenPayload{}

	err := parser.ParseOption(ctx.IC.Interaction, payload)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "parse param", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeInvalidPayload, "", err)
	}

	err = ctx.Reply(text.OperationProcessing, false)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}

	req := &controller_pb.AddAssetReq{
		ChainType:       payload.ChainType,
		ChainId:         pconst.GetChainId(payload.ChainType),
		Address:         ctx.Requester.RequesterDefaultAddress,
		TokenType:       pconst.TokenTypeErc20,
		ContractAddress: payload.ContractAddress,
	}

	transactionResp, err := ctx.CM.AddAsset(ctx.Context, req)
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if transactionResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(transactionResp.CommonResponse)
	}

	_, err = ctx.EditReply(text.OperationSuccess)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}

	return nil
}
