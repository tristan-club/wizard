package cmd_menu

import (
	"fmt"
	"github.com/tristan-club/bot-wizard/cmd"
	"github.com/tristan-club/bot-wizard/handler/text"
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/bot-wizard/handler/tghandler/tcontext"
)

var Handler = chain.NewChainHandler(cmd.CmdMenu, menuSendHandler)

func menuSendHandler(ctx *tcontext.Context) error {

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
	if _, herr := ctx.Send(ctx.U.FromChat().ID, content, nil, false, false); herr != nil {
		return herr
	}
	return nil

}
