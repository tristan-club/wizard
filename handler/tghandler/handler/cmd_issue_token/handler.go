package cmd_issue_token

import (
	"context"
	"fmt"
	"github.com/tristan-club/kit/chain_info"
	he "github.com/tristan-club/kit/error"
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
	"github.com/tristan-club/wizard/pkg/mdparse"
	"strconv"
)

const (
	initialSupplyMin = 0
	initialSupplyMax = 10000000000000
)

type IssueTokenPayload struct {
	UserNo        string `json:"user_no"`
	From          string `json:"from"`
	ChainType     uint32 `json:"chain_type"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	InitialSupply int64  `json:"initial_supply"`
	Mintable      bool   `json:"mintable"`
	PinCode       string `json:"pin_code"`
}

var Handler *chain.ChainHandler

var enterMintAble *chain.Node

func init() {

	enterNameNode := chain.NewNode(
		func(ctx *tcontext.Context, node *chain.Node) error {
			msg, herr := ctx.Send(ctx.U.SentFrom().ID, text.EnterTokenName, nil, false, false)
			if herr != nil {
				return herr
			} else {
				expiremessage_state.AddExpireMessage(ctx.OpenId(), msg)
				return nil
			}
		},
		prechecker.MustBeMessage,
		func(ctx *tcontext.Context, node *chain.Node) error {
			userstate.SetParam(ctx.OpenId(), "name", ctx.U.Message.Text)
			return nil
		})

	enterSymbolNode := chain.NewNode(
		func(ctx *tcontext.Context, node *chain.Node) error {
			msg, herr := ctx.Send(ctx.U.SentFrom().ID, text.EnterTokenSymbol, nil, false, false)
			if herr != nil {
				return herr
			} else {
				expiremessage_state.AddExpireMessage(ctx.OpenId(), msg)
				return nil
			}
		},
		prechecker.MustBeMessage,
		func(ctx *tcontext.Context, node *chain.Node) error {
			userstate.SetParam(ctx.OpenId(), "symbol", ctx.U.Message.Text)
			return nil
		})

	enterMintAble = &chain.Node{}
	*enterMintAble = *presetnode.EnterTypeNode

	Handler = chain.NewChainHandler(cmd.CmdIssueToken, issueTokenHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.SelectChainNode, nil).
		AddPresetNode(enterNameNode, nil).
		AddPresetNode(enterSymbolNode, nil).
		AddPresetNode(presetnode.EnterQuantityNode, &presetnode.EnterQuantityParam{
			Min:      initialSupplyMin,
			Max:      initialSupplyMax,
			Content:  fmt.Sprintf(text.EnterInitialSupply, initialSupplyMin, initialSupplyMax),
			ParamKey: "initial_supply",
		}).
		AddPresetNode(presetnode.EnterBoolNode, &presetnode.EnterBoolParam{
			Content:  text.EnterMintable,
			ParamKey: "mintable",
		}).
		AddPresetNode(presetnode.EnterPinCodeHandler, nil)
}

func issueTokenHandler(ctx *tcontext.Context) error {

	var payload = &IssueTokenPayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	req := &controller_pb.IssueTokenReq{
		ChainType:                payload.ChainType,
		ChainId:                  pconst.GetChainId(payload.ChainType),
		Address:                  payload.From,
		FromId:                   payload.UserNo,
		PinCode:                  payload.PinCode,
		Name:                     payload.Name,
		Symbol:                   payload.Symbol,
		InitialSupplyUnHandleStr: strconv.FormatInt(payload.InitialSupply, 10),
		MintAble:                 payload.Mintable,
		TokenType:                pconst.TokenTypeErc20,
		IsWait:                   false,
	}

	transactionResp, err := ctx.CM.IssueToken(ctx.Context, req)
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if transactionResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(transactionResp.CommonResponse)
	}

	thisMsg, herr := ctx.Send(ctx.U.FromChat().ID, fmt.Sprintf(text.TransactionProcessing, mdparse.ParseV2(pconst.GetExplore(payload.ChainType, transactionResp.Data.TxHash, chain_info.ExplorerTargetTransaction))), nil, true, false)
	if herr != nil {
		return herr
	}

	getDataResp, err := ctx.CM.GetTx(context.Background(), &controller_pb.GetTxReq{TxHash: transactionResp.Data.TxHash, IsWait: true})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if getDataResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(getDataResp.CommonResponse)
	}

	if herr := ctx.EditMessageAndKeyboard(ctx.U.FromChat().ID, thisMsg.MessageID, fmt.Sprintf(text.IssueTokenSuccess, getDataResp.Data.ContractAddress,
		mdparse.ParseV2(pconst.GetExplore(payload.ChainType, getDataResp.Data.ContractAddress, chain_info.ExplorerTargetAddress))), nil, true, false); herr != nil {
		return herr
	}

	return nil
}
