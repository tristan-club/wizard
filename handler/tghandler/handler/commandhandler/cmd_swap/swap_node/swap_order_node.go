package swap_node

import (
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/inline_keybord"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"strconv"
)

var ShowOrderNode = chain.NewNode(askForConfirm, prechecker.MustBeCallback, confirmOrCancel)

func askForConfirm(ctx *tcontext.Context, node *chain.Node) error {

	_, herr := ctx.Send(ctx.U.SentFrom().ID, text.SelectChain, &inline_keybord.ChainKeyboard, true, false)
	if herr != nil {
		return herr
	}
	return nil
}

func confirmOrCancel(ctx *tcontext.Context, node *chain.Node) error {
	chainValue := ctx.U.CallbackData()
	chainValueInt, _ := strconv.ParseInt(chainValue, 10, 64)
	userstate.SetParam(ctx.OpenId(), "chain_type", chainValueInt)
	return nil
}
