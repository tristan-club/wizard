package presetnode

import (
	"github.com/tristan-club/kit/mdparse"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
)

var EnterTextHandler = chain.NewNode(AskForText, prechecker.MustBeMessage, EnterText)

type EnterTextParam struct {
	Content  string `json:"content"`
	ParamKey string `json:"param_key"`
}

func AskForText(ctx *tcontext.Context, node *chain.Node) error {

	var param = &EnterTextParam{}

	herr := node.TryGetPayload(param)
	if herr != nil {
		return herr
	}

	if msg, herr := ctx.Send(ctx.U.SentFrom().ID, mdparse.ParseV2(param.Content), nil, true, false); herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), msg)
	}

	return nil
}

func EnterText(ctx *tcontext.Context, node *chain.Node) error {
	var param = &EnterTextParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	paramKey := param.ParamKey
	if paramKey == "" {
		paramKey = "text"
	}
	text := ctx.U.Message.Text
	userstate.SetParam(ctx.OpenId(), paramKey, text)
	return nil
}
