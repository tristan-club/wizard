package presetnode

import (
	"fmt"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pconst"
	"strconv"
)

var EnterQuantityNode = chain.NewNode(AskForQuantity, prechecker.MustBeMessage, EnterQuantity)

type EnterQuantityParam struct {
	Min      int64  `json:"min"`
	Max      int64  `json:"max"`
	Content  string `json:"content"`
	ParamKey string `json:"param_key"`
}

func AskForQuantity(ctx *tcontext.Context, node *chain.Node) error {

	var param = &EnterQuantityParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	var content string

	if param.Min == 0 || param.Max == 0 {
		if param.Content == "" {
			content = text.EnterQuantity
		} else {
			content = param.Content
		}

	} else {
		if param.Content == "" {
			content = fmt.Sprintf(text.EnterQuantityWithRange, param.Min, param.Max)
		} else {
			content = fmt.Sprintf(param.Content, param.Min, param.Max)
		}
	}
	thisMsg, herr := ctx.Send(ctx.U.SentFrom().ID, content, nil, false, false)
	if herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), thisMsg)
	}

	return nil
}

func EnterQuantity(ctx *tcontext.Context, node *chain.Node) error {

	quantity, err := strconv.ParseInt(ctx.U.Message.Text, 10, 64)
	if err != nil {
		return he.NewServerError(pconst.CodeInvalidQuantity, "", err)
	}

	var param = &EnterQuantityParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	if quantity < param.Min {
		return he.NewBusinessError(pconst.CodeInvalidQuantity, "", nil)
	}
	if param.Max != 0 {
		if quantity > param.Max {
			return he.NewBusinessError(pconst.CodeInvalidQuantity, "", nil)
		}
	}

	paramKey := param.ParamKey
	if paramKey == "" {
		paramKey = "quantity"
	}
	userstate.SetParam(ctx.OpenId(), paramKey, quantity)
	return nil

}
