package cmd_get_account

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/text"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/log"
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
	//_, herr := ctx.ReplyDmWithGroupForward(ctx.U.SentFrom().ID, fmt.Sprintf(text.GetAccountSuccess, ctx.Requester.RequesterDefaultAddress), nil, true, false)
	//return herr

	if err := ctx.Reply(fmt.Sprintf(text.GetAccountSuccess, ctx.Requester.RequesterDefaultAddress), false); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}

	return nil
}
