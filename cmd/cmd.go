package cmd

import "github.com/tristan-club/wizard/config"

const (
	CmdStart           = "start"
	CmdChangePinCode   = "change_pin_code"
	CmdGetAccount      = "get_wallet_address"
	CmdBalance         = "balance"
	CmdTransfer        = "transfer"
	CmdCreateEnvelope  = "create_red_envelope"
	CmdOpenEnvelope    = "open_red_envelope"
	CmdAddTokenBalance = "add_token_balance"
	CmdIssueToken      = "issue_token"
	CmdAirdrop         = "airdrop"
	CmdSwap            = "swap"
	CmdBridge          = "bridge"
	CmdMenu            = "menu"
	CmdMyWallet        = "my_wallet"
	CmdExportPrivate   = "export_private"
	CmdReplacePrivate  = "replace_private"
)

var cmdList = []string{
	CmdMenu, CmdStart, CmdChangePinCode, CmdGetAccount, CmdBalance, CmdTransfer, CmdCreateEnvelope, CmdOpenEnvelope, CmdAddTokenBalance, CmdIssueToken,
	CmdAirdrop, CmdSwap, CmdBridge, CmdMyWallet, CmdExportPrivate, CmdReplacePrivate,
}
var betaCmdList = []string{}

func GetCmdList() []string {
	if config.EnvIsDev() {
		cmdListCopy := make([]string, 0)
		cmdListCopy = append(cmdListCopy, cmdList...)
		cmdListCopy = append(cmdList, betaCmdList...)
		return cmdListCopy
	}
	return cmdList
}

var useWizardCmdList = []string{}

func GetUseWizardCmdList() []string {
	return useWizardCmdList
}

func SetUseWizardCmdList(cmdList []string) {
	useWizardCmdList = cmdList
}

var desc = map[string]string{
	CmdStart:           "Create your MetaWallet and get the user guide.",
	CmdChangePinCode:   "Change pin code of your MetaWallet address.",
	CmdGetAccount:      "Check your MetaWallet address.",
	CmdBalance:         "Get details of your MetaWallet balance for following assets: Crypto and NFTs",
	CmdCreateEnvelope:  "Create Red Envelopes to share with your community\n\U0001F9E7People who clicks open button can open the Red Envelope and receive tokens.",
	CmdTransfer:        "Transfer assets to certain address.",
	CmdOpenEnvelope:    "Open Red Packet shared with the community . Please specify the serial number",
	CmdAddTokenBalance: "Add specific token to display under \"/balance\" command",
	CmdIssueToken:      "Issue token with MetaWallet.",
	CmdAirdrop:         "Airdrop tokens to all community members with MetaWallet address",
	CmdSwap:            "Swap and bridge asset",
	CmdBridge:          "Bridge asset",
	CmdMenu:            "Read the menu",
	CmdMyWallet:        "Your wallet tristan uri",
	CmdReplacePrivate:  "Import a new private key to REPLACE the old key, the old key WILL NOT be recovered!",
	CmdExportPrivate:   "Export private key",
}

func GetCmdDescMap() map[string]string {
	return desc
}

func GetCmdDesc(cmd string) string {
	return desc[cmd]
}
