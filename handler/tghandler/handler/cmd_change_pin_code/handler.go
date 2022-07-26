package cmd_change_pin_code

import (
	"github.com/tristan-club/bot-wizard/cmd"
	"github.com/tristan-club/bot-wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/bot-wizard/handler/text"
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/bot-wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/bot-wizard/handler/userstate"
	he "github.com/tristan-club/bot-wizard/pkg/error"
)

type ChangePinCodePayload struct {
	UserNo     string `json:"user_no"`
	From       string `json:"from"`
	OldPinCode string `json:"old_pin_code"`
	NewPinCode string `json:"new_pin_code"`
}

var Handler *chain.ChainHandler

func init() {
	enterNewPinCodeNode := chain.NewNode(presetnode.AskForPinCode, prechecker.MustBeMessage, presetnode.EnterPinCode)

	Handler = chain.NewChainHandler(cmd.CmdChangePinCode, ChangePinCodeSendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.EnterPinCodeHandler, &presetnode.EnterPinCodeParam{
			Content:  text.EnterOldPinCode,
			ParamKey: "old_pin_code",
		}).
		AddPresetNode(enterNewPinCodeNode, &presetnode.EnterPinCodeParam{
			Content:  text.EnterNewPinCode,
			ParamKey: "new_pin_code"})

}

func ChangePinCodeSendHandler(ctx *tcontext.Context) error {

	var payload = &ChangePinCodePayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	if payload.OldPinCode == payload.NewPinCode {
		return he.NewBusinessError(he.CodeSamePinCode, "", nil)
	}

	accountResp, err := ctx.CM.ChangeAccountPinCode(ctx.Context, &controller_pb.ChangeAccountPinCodeReq{
		Address:    payload.From,
		OldPinCode: payload.OldPinCode,
		NewPinCode: payload.NewPinCode,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if accountResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(accountResp.CommonResponse)
	}

	_, herr = ctx.Send(ctx.U.SentFrom().ID, text.ChangePinCodeSuccess, nil, false, false)
	if herr != nil {
		return herr
	}
	return nil
}
