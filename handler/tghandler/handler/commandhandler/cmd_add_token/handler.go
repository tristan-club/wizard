package cmd_add_token

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

var Handler *chain.ChainHandler

type AddTokenPayload struct {
	UserNo          string `json:"user_no"`
	From            string `json:"from"`
	ChainType       uint32 `json:"chain_type"`
	TokenType       uint32 `json:"token_type"`
	ContractAddress string `json:"contract_address"`
}

var tokenTypeChoiceText = []string{"ERC20", "ERC721"}
var tokenTypeChoiceValue = []int64{pconst.TokenTypeErc20, pconst.TokenTypeERC721}

func init() {

	Handler = chain.NewChainHandler(cmd.CmdAddTokenBalance, addTokenSendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.SelectChainNode, nil).
		AddPresetNode(presetnode.EnterTypeNode, presetnode.EnterTypeParam{
			ChoiceText:         tokenTypeChoiceText,
			ChoiceValue:        tokenTypeChoiceValue,
			Content:            text.SelectAssetType,
			ParamKey:           "token_type",
			ChosenTextParamKey: "",
		}).
		AddPresetNode(presetnode.EnterAddressNode, presetnode.AddressParam{ParamKey: "contract_address", Content: text.EnterTokenAddress})
}

func addTokenSendHandler(ctx *tcontext.Context) error {

	var payload = &AddTokenPayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	req := &controller_pb.AddAssetReq{
		ChainId:         pconst.GetChainId(payload.ChainType),
		Address:         payload.From,
		TokenType:       uint32(payload.TokenType),
		ContractAddress: payload.ContractAddress,
	}

	transactionResp, err := ctx.CM.AddAsset(ctx.Context, req)
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if transactionResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(transactionResp.CommonResponse)
	}

	if _, herr := ctx.Send(ctx.U.SentFrom().ID, text.OperationSuccess, nil, false, false); herr != nil {
		return herr
	}

	return nil
}
