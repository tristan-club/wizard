package cmd_test

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
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

	var j int
	for {
		for i := 0; i < 49; i++ {
			_, err := ctx.Send(ctx.IC.ChannelID, fmt.Sprintf("round %d route %d", j, i))
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "batch error", "error": err.Error()}).Send()

			}
		}
	}

	return nil
}
