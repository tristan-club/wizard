package presetnode

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pkg/bignum"
	he "github.com/tristan-club/wizard/pkg/error"
)

const (
	MaxAmount = "MAX_AMOUNT"
)

var EnterAmountNode = chain.NewNode(AskForAmount, nil, EnterAmount)

type AmountParam struct {
	Min           string `json:"min"` // min max 暂时不兼容 withMaxButton，后面考虑拆分转账和发红包按钮
	Max           string `json:"max"`
	WithMaxButton bool   `json:"with_max_button"`
	CheckBalance  bool   `json:"check_balance"`
	Content       string `json:"content"`
	ParamKey      string `json:"param_key"`
	ShowOrder     bool   `json:"show_order"`
}

func AskForAmount(ctx *tcontext.Context, node *chain.Node) error {

	var param = &AmountParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	var content string

	if param.Min == "" || param.Max == "" {
		if param.Content == "" {
			content = text.EnterAmount
		} else {
			content = param.Content
		}

	} else {
		if param.Content == "" {
			content = fmt.Sprintf(text.EnterAmountWithRange, param.Min, param.Max)
		} else {
			content = fmt.Sprintf(param.Content, param.Min, param.Max)
		}
	}

	var ikm *tgbotapi.InlineKeyboardMarkup
	//var deadlineTime time.Duration
	//if param.WithMaxButton {
	//	ikm, deadlineTime = inline_keybord.NewMaxAmountKeyboard()
	//}

	if msg, herr := ctx.Send(ctx.U.SentFrom().ID, content, ikm, false, false); herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), msg)
		//inline_keybord.DeleteDeadKeyboard(ctx, deadlineTime, msg)
	}
	return nil
}

func EnterAmount(ctx *tcontext.Context, node *chain.Node) error {

	var param = &AmountParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}

	var amount string

	if ctx.U.Message != nil {
		amount = ctx.U.Message.Text
		amountBig, ok := bignum.HandleAddDecimal(amount, 18)
		if !ok {
			return he.NewBusinessError(he.CodeAmountParamInvalid, "", nil)
		}

		if param.Min != "" || param.Max != "" {
			min, ok := bignum.HandleAddDecimal(param.Min, 18)
			if !ok {
				log.Error().Msgf("preset amount param min %s invalid, node id %s", param.Min, node.Id)
			}
			max, ok := bignum.HandleAddDecimal(param.Max, 18)
			if !ok {
				log.Error().Msgf("preset amount param max %s invalid, node id %s", param.Max, node.Id)
			}
			if amountBig.Cmp(min) < 0 || amountBig.Cmp(max) > 0 {
				return he.NewBusinessError(he.CodeAmountParamInvalid, "", nil)
			}
		}

	} else if ctx.U.CallbackQuery != nil {
		if ctx.U.CallbackData() == MaxAmount {
			amount = MaxAmount
		}
	} else {
		log.Error().Fields(map[string]interface{}{"action": "enter amount invalid state", "message": ctx}).Send()
		return he.NewServerError(he.CodeInvalidUserState, "", fmt.Errorf("invalid user state when enter amount"))
	}

	if param.CheckBalance {
		//chainType, herr := userstate.MustUInt64(ctx.OpenId(), "chain_type")
		//if herr != nil {
		//	return herr
		//}
		//assetListResp, err := ctx.Wb.WalletMgr.AssetList(ctx.Context, &controller_pb.AssetListReq{
		//	ChainType: uint32(chainType),
		//	ChainId:   pconst.GetChainId(uint32(chainType)),
		//	Address:   ctx.Requester.RequesterDefaultAddress,
		//})
		//if err != nil {
		//	return he.NewServerError(he.CodeWalletRequestError, err)
		//} else if assetListResp.CommonResponse.Code != he.Success {
		//	return tcontext.RespToError(assetListResp.CommonResponse)
		//} else if len(assetListResp.Data.List) == 0 {
		//	return he.NewServerError(he.CodeWalletRequestError, fmt.Errorf("invalid asset response"))
		//}
		//
		//userBalance, _ := bignum.HandleAddDecimal(assetListResp.Data.List[0].BalanceCutDecimal, 18)
		//if amountBig.Cmp(userBalance) > 0 {
		//	return he.NewBusinessError(he.CodeInsufficientBalance, "")
		//}

	}

	paramKey := param.ParamKey
	if paramKey == "" {
		paramKey = "amount"
	}
	userstate.SetParam(ctx.OpenId(), paramKey, amount)
	return nil

}
