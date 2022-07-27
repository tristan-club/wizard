package presetnode

import (
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/util"
)

type AddressParam struct {
	Content  string `json:"content"`
	ParamKey string `json:"param_key"`
}

var EnterAddressNode = chain.NewNode(askForAddress, prechecker.MustBeMessage, enterAddress)

func askForAddress(ctx *tcontext.Context, node *chain.Node) error {
	var param = &AddressParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	content := param.Content
	if content == "" {
		content = text.EnterReceiverAddress
	}
	msg, herr := ctx.Send(ctx.U.SentFrom().ID, content, nil, false, false)
	if herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), msg)
	}
	return nil
}

func enterAddress(ctx *tcontext.Context, node *chain.Node) error {
	address := ctx.U.Message.Text

	addressChecked, err := util.EIP55Checksum(address)
	if err != nil {
		return he.NewServerError(he.CodeAddressParamInvalid, "", err)
	}

	var param = &AddressParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}

	paramKey := "to"
	if param.ParamKey != "" {
		paramKey = param.ParamKey
	}

	addressChecked, err = util.EIP55Checksum(address)
	if err != nil {
		return he.NewBusinessError(he.CodeAddressParamInvalid, "", nil)
	} else if len(addressChecked) != 42 {
		return he.NewBusinessError(he.CodeAddressParamInvalid, "", nil)
	}
	userstate.SetParam(ctx.OpenId(), paramKey, addressChecked)

	return nil

}
