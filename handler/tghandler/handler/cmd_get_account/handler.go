package cmd_get_account

import (
	"fmt"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
)

var Handler = chain.NewChainHandler(cmd.CmdGetAccount, getWalletAddressSendHandler).AddPreHandler(prehandler.ForwardPrivate)

func getWalletAddressSendHandler(ctx *tcontext.Context) error {
	_, herr := ctx.Send(ctx.U.SentFrom().ID, fmt.Sprintf(text.GetAccountSuccess, ctx.Requester.RequesterDefaultAddress), nil, true, false)
	return herr
}
