package presetnode

import (
	"fmt"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/inline_keybord"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	pconst2 "github.com/tristan-club/wizard/pconst"
	"strconv"
)

var SelectChainNode = chain.NewNode(askForChain, prechecker.MustBeCallback, EnterChain)

func askForChain(ctx *tcontext.Context, node *chain.Node) error {

	thisMsg, herr := ctx.Send(ctx.U.SentFrom().ID, text.SelectChain, inline_keybord.GetChainKeyBoard(ctx.Requester.RequesterAppId), false, false)
	if herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), thisMsg)
	}
	ctx.SetDeadlineMsg(ctx.U.SentFrom().ID, thisMsg.MessageID, pconst2.COMMON_KEYBOARD_DEADLINE)
	return nil
}

func EnterChain(ctx *tcontext.Context, node *chain.Node) error {
	chainValue := ctx.U.CallbackData()
	chainValueInt, _ := strconv.ParseInt(chainValue, 10, 64)
	userstate.SetParam(ctx.OpenId(), "chain_type", chainValueInt)
	//if herr := ctx.DeleteMessage(ctx.U.FromChat().ID, ctx.U.CallbackQuery.Message.MessageID); herr != nil {
	//	return herr
	//}
	chainName := pconst2.GetChainName(uint32(chainValueInt))
	if herr := ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, ctx.U.CallbackQuery.Message.MessageID, fmt.Sprintf(text.ChosenChain, chainName), nil, false, false); herr != nil {
		return herr
	}
	ctx.RemoveDeadlineMsg(ctx.U.CallbackQuery.Message.MessageID)
	return nil
}
