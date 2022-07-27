package inline_keybord

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"strconv"
	"time"
)

var ChainKeyboard = tgbotapi.InlineKeyboardMarkup{}

func init() {
	ikb := []tgbotapi.InlineKeyboardButton{}
	chainTypeList := pconst.ChainTypeList
	if config.EnvIsDev() {
		chainTypeList = pconst.DebugChainTypeList
	}
	for _, chainType := range chainTypeList {
		ikb = append(ikb, tgbotapi.NewInlineKeyboardButtonData(pconst.GetChainName(uint32(chainType)), strconv.Itoa(int(chainType))))
	}
	ChainKeyboard = tgbotapi.NewInlineKeyboardMarkup(ikb)
}

var EnvelopeTypeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("Average Amount Red Envelope", "1"),
		tgbotapi.NewInlineKeyboardButtonData("Random Amount Red Envelope", "2")},
)

func NewMaxAmountKeyboard() (*tgbotapi.InlineKeyboardMarkup, time.Duration) {
	km := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(pconst.MaxAmount, pconst.MaxAmount)})
	return &km, pconst.COMMON_KEYBOARD_DEADLINE
}

//forward keyboard need to be delete after some minutes
func NewForwardPrivateKeyBoard(ctx *tcontext.Context) (*tgbotapi.InlineKeyboardMarkup, time.Duration) {
	km := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL(text.ButtonForwardPrivateChat, fmt.Sprintf("https://t.me/%s", ctx.BotName))})
	return &km, pconst.COMMON_KEYBOARD_DEADLINE
}

func NewForwardCreateKeyBoard(ctx *tcontext.Context) (*tgbotapi.InlineKeyboardMarkup, time.Duration) {
	km := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL(text.ButtonForwardCreate, fmt.Sprintf("https://t.me/%s", ctx.BotName))})
	return &km, pconst.COMMON_KEYBOARD_DEADLINE
}

//delete keyboard when deadline coming
//do nothing if deadline is zero
func DeleteDeadKeyboard(ctx *tcontext.Context, deadline time.Duration, msg *tgbotapi.Message) {
	if msg == nil || msg.Chat == nil || deadline == 0 {
		return
	}
	go func() {
		time.Sleep(deadline)
		ctx.DeleteMessage(msg.Chat.ID, msg.MessageID)
	}()

}
