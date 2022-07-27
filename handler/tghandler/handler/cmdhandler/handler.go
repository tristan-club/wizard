package cmdhandler

import (
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/handler/tghandler/flow"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_add_token"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_airdrop"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_balance"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_bridge"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_change_pin_code"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_create_envelope"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_get_account"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_issue_token"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_menu"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_my_wallet"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_open_envelope"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_start"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_swap"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_transfer"
)

var handlerMap map[string]flow.TGFlowHandler

func init() {
	handlerMap = map[string]flow.TGFlowHandler{
		cmd.CmdMenu:            cmd_menu.Handler,
		cmd.CmdStart:           cmd_start.Handler,
		cmd.CmdGetAccount:      cmd_get_account.Handler,
		cmd.CmdChangePinCode:   cmd_change_pin_code.Handler,
		cmd.CmdBalance:         cmd_balance.Handler,
		cmd.CmdTransfer:        cmd_transfer.Handler,
		cmd.CmdCreateEnvelope:  cmd_create_envelope.Handler,
		cmd.CmdOpenEnvelope:    cmd_open_envelope.Handler,
		cmd.CmdAddTokenBalance: cmd_add_token.Handler,
		cmd.CmdIssueToken:      cmd_issue_token.Handler,
		cmd.CmdAirdrop:         cmd_airdrop.Handler,
		cmd.CmdSwap:            cmd_swap.Handler,
		cmd.CmdBridge:          cmd_bridge.Handler,
		cmd.CmdMyWallet:        cmd_my_wallet.Handler,
	}
}

func GetCmdHandler(cmdId string) flow.TGFlowHandler {
	return handlerMap[cmdId]
}
