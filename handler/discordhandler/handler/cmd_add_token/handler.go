package cmd_add_token

import (
	"github.com/bwmarrin/discordgo"
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
			presetnode.GetAddressOption(&presetnode.OptionAddressPayload{Name: "contract_address", Required: true}),
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
		return he.NewServerError(pconst.CodeInvalidPayload, "", err)
	}

	_, err = ctx.FollowUpReply(text.OperationProcessing)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
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
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if transactionResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(transactionResp.CommonResponse)
	}

	_, err = ctx.FollowUpReply(text.OperationSuccess)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	return nil
}
