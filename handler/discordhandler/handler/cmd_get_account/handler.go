package cmd_get_account

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/pconst"
)

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options:       nil,
		Version:       "1",
	},
	Handler: getWalletAddressSendHandler,
}

func getWalletAddressSendHandler(ctx *dcontext.Context) error {

	if _, err := ctx.FollowUpReply(fmt.Sprintf(text.GetAccountSuccess, ctx.Requester.RequesterDefaultAddress)); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	return nil
}
