package cmd_export_key

import (
	"fmt"
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

type ExportKeyPayload struct {
	UserNo  string `json:"user_no"`
	From    string `json:"from"`
	PinCode string `json:"pin_code"`
}

var Handler *chain.ChainHandler

func init() {
	Handler = chain.NewChainHandler(cmd.CmdChangePinCode, ExportKeySendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.EnterPinCodeHandler, nil)
}

func ExportKeySendHandler(ctx *tcontext.Context) error {

	var payload = &ExportKeyPayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	accountResp, err := ctx.CM.GetAccount(ctx.Context, &controller_pb.GetAccountReq{
		UserNo:  payload.UserNo,
		PinCode: payload.PinCode,
	})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if accountResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(accountResp.CommonResponse)
	}

	_, herr = ctx.Send(ctx.U.SentFrom().ID, fmt.Sprintf(text.GetPrivateSuccessNeedDeleteMsg, accountResp.Data.PrivateKey), nil, true, false)
	if herr != nil {
		return herr
	}
	return nil
}
