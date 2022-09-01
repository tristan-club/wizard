package cmd_menu

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
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
	Handler: menuSendHandler,
}

func menuSendHandler(ctx *dcontext.Context) error {
	cmdDesc := "⚙️ Commands\n"

	cmdList := cmd.GetUseWizardCmdList()
	if len(cmdList) == 0 {
		cmdList = cmd.GetCmdList()
	}
	for _, v := range cmdList {
		cmdDesc += fmt.Sprintf("/%s %s\n", v, cmd.GetCmdDesc(v))
	}
	content := "ℹ️ User Guide\n"
	content += text.Introduce

	content += "\n"
	content += "\n"
	content += cmdDesc
	content += "\n"
	if _, err := ctx.FollowUpReply(content); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}
	return nil

}
