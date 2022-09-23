package cmd_submit_metamask

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
)

type BindMetamaskPayload struct {
	UserNo  string `json:"user_no"`
	Address string `json:"address"`
}

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options: []*discordgo.ApplicationCommandOption{
			presetnode.GetAddressOption(&presetnode.OptionAddressPayload{Name: "address", Required: true, Description: "Please enter your MetaMask address"}),
		},
		Version: "1",
	},
	Handler: submitMetaMask,
}

func submitMetaMask(ctx *dcontext.Context) error {

	var payload = &BindMetamaskPayload{}
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

	req := &controller_pb.UpdateUserReq{
		UserNo:          ctx.Requester.RequesterUserNo,
		MetamaskAddress: payload.Address,
	}

	resp, err := ctx.CM.UpdateUser(ctx.Context, req)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "call controller error", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if resp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "update user error", "error": resp}).Send()
		return tcontext.RespToError(resp.CommonResponse)
	}

	_, err = ctx.FollowUpReply(fmt.Sprintf(text.BindMetamaskAddressSuccess, payload.Address))
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	return nil
}
