package cmdhandler

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_add_token"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_balance"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_change_pincode"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_export_key"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_get_account"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_import_key"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_menu"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_my_wallet"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_start"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_transfer"
)

var handlerMap map[string]*handler.DiscordCmdHandler

func init() {
	handlerMap = map[string]*handler.DiscordCmdHandler{
		cmd.CmdMenu:          cmd_menu.Handler,
		cmd.CmdStart:         cmd_start.Handler,
		cmd.CmdGetAccount:    cmd_get_account.Handler,
		cmd.CmdChangePinCode: cmd_change_pincode.Handler,
		cmd.CmdBalance:       cmd_balance.Handler,
		cmd.CmdTransfer:      cmd_transfer.Handler,
		//cmd.CmdCreateEnvelope:  cmd_create_envelope.Handler,
		//cmd.CmdOpenEnvelope:    cmd_open_envelope.Handler,
		cmd.CmdAddTokenBalance: cmd_add_token.Handler,
		//cmd.CmdIssueToken:      cmd_issue_token.Handler,
		//cmd.CmdAirdrop:         cmd_airdrop.Handler,
		//cmd.CmdSwap:            cmd_swap.Handler,
		//cmd.CmdBridge:          cmd_bridge.Handler,
		cmd.CmdMyWallet:       cmd_my_wallet.Handler,
		cmd.CmdReplacePrivate: cmd_import_key.Handler,
		cmd.CmdExportPrivate:  cmd_export_key.Handler,
	}

	for k, _ := range handlerMap {
		handlerMap[k].ApplicationCommand.Type = discordgo.ChatApplicationCommand
		handlerMap[k].ApplicationCommand.Name = k
		handlerMap[k].ApplicationCommand.Description = cmd.GetCmdDesc(k)
		//handlerMap[k].ApplicationCommand.ID = k

		if handlerMap[k].ApplicationCommand.Version == "" {
			handlerMap[k].ApplicationCommand.Version = "1"
		}

		// 描述长度>100时discord api会报错
		if k == cmd.CmdCreateEnvelope {
			handlerMap[k].ApplicationCommand.Description = "Create Red Envelopes to share with your community"
		}
	}

}

func GetCmdHandler(cmdId string) *handler.DiscordCmdHandler {
	return handlerMap[cmdId]
}
