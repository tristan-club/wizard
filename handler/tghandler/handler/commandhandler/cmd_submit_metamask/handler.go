package cmd_submit_metamask

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pconst"
)

var Handler *chain.ChainHandler

type BindMetamaskPayload struct {
	UserNo  string `json:"user_no"`
	From    string `json:"from"`
	Address string `json:"address"`
}

func init() {

	Handler = chain.NewChainHandler(cmd.CmdAddTokenBalance, subMetaMask).
		AddCmdParser(func(u *tgbotapi.Update) string {
			if u.CallbackData() == cmd.CmdSubmitMetamask {
				return cmd.CmdSubmitMetamask
			}
			return ""
		}).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(chain.NewNode(askForMetamaskAddress, prechecker.MustBeMessage, presetnode.EnterAddress), &presetnode.AddressParam{ParamKey: "address"})
}

func askForMetamaskAddress(ctx *tcontext.Context, node *chain.Node) error {

	var content string
	if ctx.Requester.MetamaskAddress == "" {
		content = text.NoMetamaskAddress
	} else {
		content = fmt.Sprintf(text.HasMetamaskAddress, ctx.Requester.MetamaskAddress)
	}

	msg, herr := ctx.Send(ctx.U.SentFrom().ID, content, nil, true, false)
	if herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), msg)
	}
	return nil
}

func subMetaMask(ctx *tcontext.Context) error {

	var payload = &BindMetamaskPayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	req := &controller_pb.UpdateUserReq{
		UserNo:          ctx.Requester.RequesterUserNo,
		MetamaskAddress: payload.Address,
	}

	resp, err := ctx.CM.UpdateUser(ctx.Context, req)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "call controller error", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if resp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "update user error", "error": resp}).Send()
		return tcontext.RespToError(resp.CommonResponse)
	}

	if _, herr = ctx.Send(ctx.U.SentFrom().ID, fmt.Sprintf(text.BindMetamaskAddressSuccess, payload.Address), nil, true, false); herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg error", "error": herr}).Send()
		return herr
	}

	return nil
}
