package cmd_menu

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/bot-wizard/cmd"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/handler"
	"github.com/tristan-club/bot-wizard/handler/text"
	he "github.com/tristan-club/bot-wizard/pkg/error"
	"github.com/tristan-club/bot-wizard/pkg/log"
)

var Handler = &handler.DiscordCmdHandler{
	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		Options:       nil,
		Version:       "1",
	},
	Handler: menuSendHandler,
}

func menuSendHandler(ctx *dcontext.Context) error {

	cmdDesc := "⚙️ Commands\n"
	for _, v := range cmd.GetCmdList() {
		cmdDesc += fmt.Sprintf("/%s %s\n", v, cmd.GetCmdDesc(v))
	}
	content := "ℹ️ User Guide\n"
	content += text.Introduce

	content += "\n"
	content += "\n"
	content += cmdDesc
	content += "\n"
	if err := ctx.Reply(content); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}
	return nil

}
