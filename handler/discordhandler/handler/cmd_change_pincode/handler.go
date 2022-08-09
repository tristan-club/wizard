package cmd_change_pincode

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
			presetnode.GetPinCodeOption("old_pin_code", "Enter your old pin code"),
			presetnode.GetPinCodeOption("new_pin_code", "Enter your new pin code"),
		},
		Version: "1",
	},
	Handler: ChangePinCodeSendHandler,
}

type ChangePinCodePayload struct {
	OldPinCode string `json:"old_pin_code"`
	NewPinCode string `json:"new_pin_code"`
}

func ChangePinCodeSendHandler(ctx *dcontext.Context) error {

	var payload = &ChangePinCodePayload{}
	err := parser.ParseOption(ctx.IC.Interaction, payload)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "parse param", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeInvalidPayload, "", err)
	}

	if payload.OldPinCode == payload.NewPinCode {
		return he.NewBusinessError(he.CodeSamePinCode, "", nil)
	}

	accountResp, err := ctx.CM.ChangeAccountPinCode(ctx.Context, &controller_pb.ChangeAccountPinCodeReq{
		Address:    ctx.Requester.RequesterDefaultAddress,
		OldPinCode: payload.OldPinCode,
		NewPinCode: payload.NewPinCode,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if accountResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(accountResp.CommonResponse)
	}

	err = ctx.FollowUpReply(text.ChangePinCodeSuccess)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}

	return nil
}
