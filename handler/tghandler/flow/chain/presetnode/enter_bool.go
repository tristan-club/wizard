package presetnode

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pconst"
)

var EnterBoolNode = chain.NewNode(AskForBool, prechecker.MustBeCallback, EnterBool)

type EnterBoolParam struct {
	Content  string `json:"content"`
	ParamKey string `json:"param_key"`
}

func AskForBool(ctx *tcontext.Context, node *chain.Node) error {

	var param = &EnterBoolParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	ikb := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Yes", "Yes"), tgbotapi.NewInlineKeyboardButtonData("No", "No")}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(ikb)
	content := text.SelectEnvelopeRewardType
	if param.Content != "" {
		content = param.Content
	}
	thisMsg, herr := ctx.Send(ctx.U.SentFrom().ID, content, &keyboard, false, false)
	if herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), thisMsg)
	}
	ctx.SetDeadlineMsg(ctx.U.SentFrom().ID, thisMsg.MessageID, pconst.COMMON_KEYBOARD_DEADLINE)
	return nil
}

func EnterBool(ctx *tcontext.Context, node *chain.Node) error {
	var value bool
	if ctx.U.CallbackData() == "Yes" {
		value = true
	}

	var param = &EnterBoolParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}

	paramKey := "type"
	if param.ParamKey != "" {
		paramKey = param.ParamKey
	}

	userstate.SetParam(ctx.OpenId(), paramKey, value)

	//if herr := ctx.DeleteMessage(ctx.U.FromChat().ID, ctx.U.CallbackQuery.Message.MessageID); herr != nil {
	//	return herr
	//}

	if herr := ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, ctx.U.CallbackQuery.Message.MessageID, fmt.Sprintf(text.ChosenCommon, ctx.U.CallbackData()), nil, false, false); herr != nil {
		return herr
	}
	ctx.RemoveDeadlineMsg(ctx.U.CallbackQuery.Message.MessageID)
	return nil
}
