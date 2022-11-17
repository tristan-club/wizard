package cmd_menu

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
)

var Handler = chain.NewChainHandler(cmd.CmdMenu, menuSendHandler).
	AddCmdParser(func(u *tgbotapi.Update) string {
		if u.CallbackData() == cmd.CmdMenu {
			return cmd.CmdMenu
		}
		if u.Message != nil && u.Message.Text == text.KBHelp {
			return cmd.CmdMenu
		}
		return ""
	})

func menuSendHandler(ctx *tcontext.Context) error {

	var content string
	if text.CustomStartMenu != "" {
		content = text.CustomStartMenu
	} else {
		cmdDesc := "⚙️ Commands\n"
		for _, v := range cmd.GetCmdList() {
			cmdDesc += fmt.Sprintf("/%s %s\n", v, cmd.GetCmdDesc(v))
		}
		content = "ℹ️ User Guide\n"
		content += text.Introduce

		content += "\n"
		content += "\n"
		content += cmdDesc
		content += "\n"
	}

	if _, herr := ctx.Send(ctx.U.FromChat().ID, content, nil, true, false); herr != nil {
		return herr
	}
	return nil

}
