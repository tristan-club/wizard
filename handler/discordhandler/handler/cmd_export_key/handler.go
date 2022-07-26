package cmd_export_key

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
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"strings"
)

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options: []*discordgo.ApplicationCommandOption{
			presetnode.GetPinCodeOption("", ""),
		},
		Version: "1",
	},
	Handler: ImportKeyHandler,
}

type ImportKeyPayload struct {
	PinCode string `json:"pin_code"`
}

func ImportKeyHandler(ctx *dcontext.Context) error {

	var payload = &ImportKeyPayload{}
	err := parser.ParseOption(ctx.IC.Interaction, payload)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "parse param", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeInvalidPayload, "", err)
	}

	var content string

	if strings.HasPrefix(payload.PinCode, pconst.MockDeleteAccountCode) {
		resp, err := ctx.CM.DeleteAccount(ctx.Context, &controller_pb.DeleteAccountReq{
			UserNo:  ctx.Requester.RequesterUserNo,
			PinCode: strings.TrimPrefix(payload.PinCode, pconst.MockDeleteAccountCode),
		})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "request controller error", "error": err.Error()}).Send()
			return he.NewServerError(pconst.CodeWalletRequestError, "", err)
		} else if resp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "controller get account error", "error": resp}).Send()
			return tcontext.RespToError(resp.CommonResponse)
		}
		content = text.OperationSuccess
	} else {
		resp, err := ctx.CM.GetAccount(ctx.Context, &controller_pb.GetAccountReq{
			UserNo:  ctx.Requester.RequesterUserNo,
			PinCode: payload.PinCode,
		})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "request controller error", "error": err.Error()}).Send()
			return he.NewServerError(pconst.CodeWalletRequestError, "", err)
		} else if resp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "controller get account error", "error": resp}).Send()
			return tcontext.RespToError(resp.CommonResponse)
		}
		content = fmt.Sprintf(text.GetPrivateSuccess, resp.Data.PrivateKey)

	}

	_, err = ctx.FollowUpReply(content)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	return nil
}
