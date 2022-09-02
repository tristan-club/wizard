package cmd_delete_account

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
	Handler = chain.NewChainHandler(cmd.CmdDeleteAccount, DeleteAccountSendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.EnterPinCodeHandler, &presetnode.EnterPinCodeParam{Content: text.EnterPinCodeToDelete})
}

func DeleteAccountSendHandler(ctx *tcontext.Context) error {

	var payload = &ImportKeyPayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	resp, err := ctx.CM.DeleteUser(ctx.Context, &controller_pb.DeleteUserReq{
		UserNo:  payload.UserNo,
		PinCode: payload.PinCode,
	})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if resp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(resp.CommonResponse)
	}

	_, herr = ctx.Send(ctx.U.SentFrom().ID, text.OperationSuccess, nil, false, false)
	if herr != nil {
		return herr
	}
	return nil
}
