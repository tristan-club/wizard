package cmd_import_key

import (
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
)

type ImportKeyPayload struct {
	UserNo     string `json:"user_no"`
	From       string `json:"from"`
	PrivateKey string `json:"private_key"`
	PinCode    string `json:"pin_code"`
}

var Handler *chain.ChainHandler

func init() {
	Handler = chain.NewChainHandler(cmd.CmdChangePinCode, ImportKeySendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.EnterTextHandler, &presetnode.EnterTextParam{
			Content:  text.EnterPrivateWithDelete,
			ParamKey: "private_key",
		}).
		AddPresetNode(presetnode.EnterPinCodeHandler, nil)
}

func ImportKeySendHandler(ctx *tcontext.Context) error {

	var payload = &ImportKeyPayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	accountResp, err := ctx.CM.ImportAccount(ctx.Context, &controller_pb.ImportAccountReq{
		UserNo:     payload.UserNo,
		PrivateKey: payload.PrivateKey,
		PinCode:    payload.PinCode,
	})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if accountResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(accountResp.CommonResponse)
	}

	_, herr = ctx.Send(ctx.U.SentFrom().ID, text.OperationSuccess, nil, false, false)
	if herr != nil {
		return herr
	}
	return nil
}
