package cmd_import_key

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/flow/presetnode"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/discordhandler/parser"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	he "github.com/tristan-club/wizard/pkg/error"
)

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "private_key",
				Description: "Your private key",
				Required:    true,
			},
			presetnode.GetPinCodeOption("", ""),
		},
		Version: "1",
	},
	Handler: ImportKeyHandler,
}

type ExportKeyPayload struct {
	PinCode    string `json:"pin_code"`
	PrivateKey string `json:"private_key"`
}

func ImportKeyHandler(ctx *dcontext.Context) error {

	var payload = &ExportKeyPayload{}
	err := parser.ParseOption(ctx.IC.Interaction, payload)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "parse param", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeInvalidPayload, "", err)
	}

	resp, err := ctx.CM.ImportAccount(ctx.Context, &controller_pb.ImportAccountReq{
		UserNo:     ctx.Requester.RequesterUserNo,
		PrivateKey: payload.PrivateKey,
		PinCode:    payload.PinCode,
	})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "request controller error", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if resp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "controller error", "error": resp}).Send()
		return tcontext.RespToError(resp.CommonResponse)
	}

	err = ctx.FollowUpReply(text.OperationSuccess)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}

	return nil
}
