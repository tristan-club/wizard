package tcontext

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/rate"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/util"
	"io"
	"net/http"
	"strings"
)

func (ctx *Context) Send(chatId int64, content string, ikm interface{}, markdownContent bool, disablePreview bool) (*tgbotapi.Message, he.Error) {
	rate.CheckLimit(chatId)
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
		return nil, he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}
	return &thisMSg, nil
}
func (ctx *Context) Reply(chatId int64, content string, ikm *tgbotapi.InlineKeyboardMarkup, markdownContent bool) (tgbotapi.Message, he.Error) {
	rate.CheckLimit(chatId)
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
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}
	return nil
}

func (ctx *Context) EditMessageAndKeyboard(chatId int64, messageId int, content string, ikm *tgbotapi.InlineKeyboardMarkup, markdownContent bool, disableWebPreview bool) he.Error {
	editMsg := tgbotapi.NewEditMessageText(chatId, messageId, content)
	if ikm != nil {
		editMsg.ReplyMarkup = ikm
	}
	if markdownContent {
		editMsg.ParseMode = tgbotapi.ModeMarkdownV2
	}
	editMsg.DisableWebPagePreview = disableWebPreview

	if _, err := ctx.BotApi.Send(editMsg); err != nil {
		log.Error().Fields(map[string]interface{}{
			"action":  "edit msg",
			"error":   err.Error(),
			"chatId":  chatId,
			"msgId":   messageId,
			"payload": util.FastMarshal(ctx),
		}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}
	return nil
}

func (ctx *Context) ReplyDmWithGroupForward(content string, ikm *tgbotapi.InlineKeyboardMarkup) he.Error {
	rate.CheckLimit(ctx.U.SentFrom().ID)
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
			return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
		}
	}
	return nil
}

func (ctx *Context) SetChatMenuButton(menuButton *tgbotapi.MenuButton) error {
	rate.CheckLimit(ctx.U.SentFrom().ID)
	req := tgbotapi.SetChatMenuButtonConfig{
		ChatID:          ctx.U.SentFrom().ID,
		ChannelUsername: "",
		MenuButton:      menuButton,
	}
	res, err := ctx.BotApi.Request(req)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send chat button error", "error": err.Error()}).Send()
		return err
	} else if !res.Ok {
		log.Error().Fields(map[string]interface{}{"action": "bot send button get error", "error": res}).Send()
		return fmt.Errorf(res.Description)
	}

	return nil
}

func (ctx *Context) SendPhoto(chatId int64, content string, ikm interface{}, markdownContent bool, photoUrl string) (*tgbotapi.Message, he.Error) {
	rate.CheckLimit(chatId)
	photoConfig := tgbotapi.NewPhoto(chatId, tgbotapi.FileURL(photoUrl))
	photoConfig.Caption = content
	photoConfig.ReplyMarkup = ikm
	if markdownContent {
		photoConfig.ParseMode = tgbotapi.ModeMarkdownV2
	}
	m, err := ctx.BotApi.Send(photoConfig)
	if err != nil {
		if isSendPhotoError(err) {
			m, err = ctx.retrySendPhoto(photoConfig, photoUrl)
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "retry send photo error", "error": err.Error(), "photoUrl": photoUrl}).Send()
			}
		} else {
			log.Error().Fields(map[string]interface{}{
				"action": "telegram bot send photo",
				"chat":   chatId,
				"name":   ctx.BotName,
				"photo":  photoUrl,
				"error":  err,
			}).Send()
		}
	}
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "retry send photo error", "error": err.Error(), "photoUrl": photoUrl}).Send()
		return nil, he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}
	return &m, nil
}

func isSendPhotoError(err error) bool {
	if err == nil {
		log.Warn().Msgf("empty error asset for send photo")
		return false
	}
	if strings.Contains(err.Error(), "wrong file identifier/HTTP URL specified") ||
		strings.Contains(err.Error(), "failed to get HTTP URL content") {
		return true
	}
	return false
}

func (ctx *Context) retrySendPhoto(pc tgbotapi.PhotoConfig, photoUrl string) (msg tgbotapi.Message, err error) {
	httpResp, err := http.Get(photoUrl)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "get photo by url error", "error": err.Error(), "pc": pc}).Send()
		return msg, err
	}
	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "ready photo content error", "error": err.Error(), "pc": pc}).Send()
		return msg, err
	}

	pc.File = tgbotapi.FileBytes{Name: "Photo", Bytes: body}
	msg, err = ctx.BotApi.Send(pc)
	//if err != nil && isSendPhotoError(err) {
	//
	//} else {
	//	log.Error().Fields(map[string]interface{}{"action": "send photo by content error", "error": err.Error(), "pc": pc}).Send()
	//	return msg, err
	//}
	log.Error().Fields(map[string]interface{}{"action": "send photo by content error", "error": err.Error(), "pc": pc}).Send()
	return msg, err

}
