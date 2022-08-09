package presetnode

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pconst"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/util"
	"strconv"
)

var EnterTypeNode = chain.NewNode(AskForType, prechecker.MustBeCallback, EnterType)

type EnterTypeParam struct {
	ChoiceText         []string `json:"choice_text"`
	ChoiceValue        []int64  `json:"choice_value"`
	Content            string   `json:"content"`
	ParamKey           string   `json:"param_key"`
	ChosenTextParamKey string   `json:"chosen_text_param_key"`
}

func AskForType(ctx *tcontext.Context, node *chain.Node) error {

	var param = &EnterTypeParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	if len(param.ChoiceText) == 0 || len(param.ChoiceValue) == 0 || len(param.ChoiceText) != len(param.ChoiceValue) {
		log.Error().Fields(map[string]interface{}{
			"action": "invalid choice ",
			"param":  util.FastMarshal(param),
			"ctx":    util.FastMarshal(ctx),
		}).Send()
		return he.NewServerError(he.CodeInvalidChoice, "", fmt.Errorf("invalid choice"))
	}
	ikb := make([]tgbotapi.InlineKeyboardButton, 0)
	for k, v := range param.ChoiceText {
		ikb = append(ikb, tgbotapi.NewInlineKeyboardButtonData(v, strconv.FormatInt(param.ChoiceValue[k], 10)))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(ikb)
	content := text.SelectEnvelopeType
	if param.Content != "" {
		content = param.Content
	}

	thisMsg, herr := ctx.Send(ctx.U.SentFrom().ID, content, &keyboard, true, false)
	if herr != nil {
		return herr
	} else {
		expiremessage_state.AddExpireMessage(ctx.OpenId(), thisMsg)
	}
	ctx.SetDeadlineMsg(ctx.U.SentFrom().ID, thisMsg.MessageID, pconst.COMMON_KEYBOARD_DEADLINE)
	return nil
}

func EnterType(ctx *tcontext.Context, node *chain.Node) error {
	typeParam, err := strconv.ParseInt(ctx.U.CallbackData(), 10, 64)
	if err != nil {
		return he.NewServerError(he.CodeInvalidType, "", err)
	}
	var param = &EnterTypeParam{}
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

	chosenText := ""
	if len(param.ChoiceText) != 0 && len(param.ChoiceValue) != 0 {
		for k, v := range param.ChoiceValue {
			if v == typeParam {
				chosenText = param.ChoiceText[k]
			}
		}
	}
	chosenTextParamKey := "chosen_text"
	if param.ChosenTextParamKey != "" {
		chosenTextParamKey = param.ChosenTextParamKey
	}
	userstate.SetParam(ctx.OpenId(), paramKey, typeParam)
	if chosenText != "" {
		userstate.SetParam(ctx.OpenId(), chosenTextParamKey, chosenText)
	}

	if chosenText != "" {
		if herr := ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, ctx.U.CallbackQuery.Message.MessageID, fmt.Sprintf(text.ChosenCommon, chosenText), nil, false, false); herr != nil {
			return herr
		}
	} else {
		if herr := ctx.DeleteMessage(ctx.U.SentFrom().ID, ctx.U.CallbackQuery.Message.MessageID); herr != nil {
			return herr
		}
	}
	ctx.RemoveDeadlineMsg(ctx.U.CallbackQuery.Message.MessageID)
	return nil
}
