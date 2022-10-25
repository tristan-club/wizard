package cmdhandler

import (
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/handler/tghandler/flow"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_add_token"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_airdrop"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_balance"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_bridge"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_change_pin_code"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_create_envelope"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_delete_account"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_export_key"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_get_account"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_import_key"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_issue_token"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_menu"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_my_wallet"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_open_envelope"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_start"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_submit_metamask"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_swap"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_test"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmd_transfer"
)

var handlerMap map[string]flow.TGFlowHandler

func init() {
	handlerMap = map[string]flow.TGFlowHandler{
		cmd.CmdMenu:            cmd_menu.Handler,
		cmd.CmdStart:           cmd_start.Handler,
		cmd.CmdGetAccount:      cmd_get_account.Handler,
		cmd.CmdChangePinCode:   cmd_change_pin_code.Handler,
		cmd.CmdSubmitMetamask:  cmd_submit_metamask.Handler,
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
		cmd.CmdExportPrivate:   cmd_export_key.Handler,
		cmd.CmdReplacePrivate:  cmd_import_key.Handler,
		cmd.CmdDeleteAccount:   cmd_delete_account.Handler,
		cmd.CmdTest:            cmd_test.Handler,
	}
}

func GetCmdHandler(cmdId string) flow.TGFlowHandler {
	return handlerMap[cmdId]
}
