package tcontext

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/handler/text"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/log"
	"github.com/tristan-club/wizard/pkg/util"
)

func (ctx *Context) Send(chatId int64, content string, ikm interface{}, markdownContent bool, disablePreview bool) (*tgbotapi.Message, he.Error) {
	var message *tgbotapi.Message
	var thisMSg tgbotapi.Message
	if ctx.U.Message != nil {
		message = ctx.U.Message
	} else if ctx.U.CallbackQuery != nil {
		message = ctx.U.CallbackQuery.Message
	} else {
		log.Error().Msgf("unknown message, chatId %d, content %s, payload %s", chatId, content, util.FastMarshal(ctx))
		return &thisMSg, nil
	}

	if chatId == 0 {
		chatId = message.Chat.ID
	}

	msg := tgbotapi.NewMessage(chatId, content)
	if ikm != nil {
		msg.ReplyMarkup = ikm
	}
	if markdownContent {
		msg.ParseMode = tgbotapi.ModeMarkdownV2
	}
	msg.DisableWebPagePreview = disablePreview
	thisMSg, err := ctx.BotApi.Send(msg)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action":   "telegram bot send message",
			"bot name": ctx.BotName,
			"chat":     chatId,
			"content":  content,
			"error":    err,
		}).Send()
		return nil, he.NewServerError(he.CodeBotSendMsgError, "", err)
	}
	return &thisMSg, nil
}
func (ctx *Context) Reply(chatId int64, content string, ikm *tgbotapi.InlineKeyboardMarkup, markdownContent bool) (tgbotapi.Message, he.Error) {
	var message *tgbotapi.Message
	var thisMsg tgbotapi.Message
	if ctx.U.Message != nil {
		message = ctx.U.Message
	} else if ctx.U.CallbackQuery != nil {
		message = ctx.U.CallbackQuery.Message
	} else {
		log.Error().Msgf("unknown message, chatId %d, content %s, payload %s", chatId, content, util.FastMarshal(ctx))
		return thisMsg, nil
	}
	if chatId == 0 {
		chatId = message.Chat.ID
	}

	reply := tgbotapi.NewMessage(message.Chat.ID, content)
	if ikm != nil {
		reply.ReplyMarkup = ikm
	}

	if markdownContent {
		reply.ParseMode = tgbotapi.ModeMarkdownV2
	}

	reply.ReplyToMessageID = message.MessageID
	m, err := ctx.BotApi.Send(reply)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action":   "telegram bot reply message",
			"bot name": ctx.BotName,
			"openId":   ctx.OpenId(),
			"error":    err,
		}).Send()
	}
	return m, nil
}

func (ctx *Context) TryDeleteMessage(message *tgbotapi.Message) he.Error {
	if message == nil {
		log.Warn().Fields(map[string]interface{}{"action": "deleteMessage", "cmd": ctx.CmdId})
		return nil
	}
	if message.Chat == nil {
		log.Warn().Fields(map[string]interface{}{"action": "deleteMessage", "ctx": ctx.CmdId})
		return nil
	}
	return ctx.DeleteMessage(message.Chat.ID, message.MessageID)
}

func (ctx *Context) DeleteMessage(chatId int64, messageId int) he.Error {
	delMsg := tgbotapi.NewDeleteMessage(chatId, messageId)
	_, err := ctx.BotApi.Request(delMsg)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action":  "delete msg",
			"error":   err.Error(),
			"chatId":  chatId,
			"msgId":   messageId,
			"payload": util.FastMarshal(ctx),
		}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}
	return nil
}

func (ctx *Context) EditMessageAndKeyboard(chatId int64, messageId int, content string, ikm *tgbotapi.InlineKeyboardMarkup, markdownContent bool, disableWebPreview bool) he.Error {
	var editMsg tgbotapi.EditMessageTextConfig
	if ikm != nil {
		editMsg = tgbotapi.NewEditMessageTextAndMarkup(chatId, messageId, content, *ikm)
	} else {
		editMsg = tgbotapi.NewEditMessageText(chatId, messageId, content)
	}
	if markdownContent {
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	}
	editMsg.DisableWebPagePreview = disableWebPreview

	if _, err := ctx.BotApi.Request(editMsg); err != nil {
		log.Error().Fields(map[string]interface{}{
			"action":  "delete msg",
			"error":   err.Error(),
			"chatId":  chatId,
			"msgId":   messageId,
			"payload": util.FastMarshal(ctx),
		}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}
	return nil
}

func (ctx *Context) ReplyDmWithGroupForward(content string, ikm *tgbotapi.InlineKeyboardMarkup) he.Error {
	var message *tgbotapi.Message
	if ctx.U.Message != nil {
		message = ctx.U.Message
	} else if ctx.U.CallbackQuery != nil {
		message = ctx.U.CallbackQuery.Message
	} else {
		log.Error().Msgf("unknown message, content %s, payload %s", content, util.FastMarshal(ctx))
		return nil
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, content)
	if ikm != nil {
		msg.ReplyMarkup = ikm
	}
	_, err := ctx.BotApi.Send(msg)

	if !message.Chat.IsPrivate() {
		_, err = ctx.Reply(0, text.CheckDm, nil, false)
		if err != nil {
			return he.NewServerError(he.CodeBotSendMsgError, "", err)
		}
	}
	return nil
}
