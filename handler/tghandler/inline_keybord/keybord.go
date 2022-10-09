package inline_keybord

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"strconv"
	"time"
)

var ChainKeyboard = tgbotapi.InlineKeyboardMarkup{}

var appAvailableChainKeyboardMap = map[string]tgbotapi.InlineKeyboardMarkup{}

func AddAppChain(appId string, chainTypeList []uint32) {

	if appId == "" || len(chainTypeList) == 0 {
		log.Error().Msgf("invalid app chain config. appId %s, chainTypeList %v", appId, chainTypeList)
		return
	}

	kb := tgbotapi.InlineKeyboardMarkup{}
	ikb := make([]tgbotapi.InlineKeyboardButton, 0)
	availableChain := make([]uint32, 0)

	for _, chainType := range chainTypeList {
		for _, v := range pconst.ChainTypeList {
			if chainType == v {
				availableChain = append(availableChain, chainType)
				break
			}
		}
	}

	if len(availableChain) == 0 {
		log.Error().Msgf("invalid app chain config. appId %s, chainTypeList %v", appId, chainTypeList)
		return
	}

	for _, chainType := range availableChain {
		ikb = append(ikb, tgbotapi.NewInlineKeyboardButtonData(pconst.GetChainName(uint32(chainType)), strconv.Itoa(int(chainType))))
	}
	if len(ikb) <= 3 {
		kb = tgbotapi.NewInlineKeyboardMarkup(ikb)
	} else {
		ikbArray := make([][]tgbotapi.InlineKeyboardButton, 0)
		for len(ikb) > 3 {
			ikbArray = append(ikbArray, ikb[:3])
			ikb = ikb[3:]
		}
		ikbArray = append(ikbArray, ikb)
		kb = tgbotapi.NewInlineKeyboardMarkup(ikbArray...)
	}

	appAvailableChainKeyboardMap[appId] = kb
}

func init() {
	ikb := []tgbotapi.InlineKeyboardButton{}
	chainTypeList := pconst.ChainTypeList
	if config.EnvIsDev() {
		chainTypeList = pconst.DebugChainTypeList
	}
	for _, chainType := range chainTypeList {
		ikb = append(ikb, tgbotapi.NewInlineKeyboardButtonData(pconst.GetChainName(uint32(chainType)), strconv.Itoa(int(chainType))))
	}
	if len(ikb) <= 3 {
		ChainKeyboard = tgbotapi.NewInlineKeyboardMarkup(ikb)
	} else {
		ikbArray := make([][]tgbotapi.InlineKeyboardButton, 0)
		for len(ikb) > 3 {
			ikbArray = append(ikbArray, ikb[:3])
			ikb = ikb[3:]
		}
		ikbArray = append(ikbArray, ikb)
		ChainKeyboard = tgbotapi.NewInlineKeyboardMarkup(ikbArray...)
	}

}

func GetChainKeyBoard(appId string) *tgbotapi.InlineKeyboardMarkup {

	if kb, ok := appAvailableChainKeyboardMap[appId]; ok {
		return &kb
	}

	return &ChainKeyboard
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

// forward keyboard need to be delete after some minutes
func NewForwardPrivateKeyBoard(ctx *tcontext.Context) (*tgbotapi.InlineKeyboardMarkup, time.Duration) {
	km := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL(text.ButtonForwardPrivateChat, fmt.Sprintf("https://t.me/%s", ctx.BotName))})
	return &km, pconst.ForwardPrivateDeadline
}

func NewForwardCreateKeyBoard(ctx *tcontext.Context) (*tgbotapi.InlineKeyboardMarkup, time.Duration) {
	km := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonURL(text.ButtonForwardCreate, fmt.Sprintf("https://t.me/%s?start=default", ctx.BotName))})
	return &km, pconst.ForwardPrivateDeadline
}

// delete keyboard when deadline coming
// do nothing if deadline is zero
func DeleteDeadKeyboard(ctx *tcontext.Context, deadline time.Duration, msg *tgbotapi.Message) {
	if msg == nil || msg.Chat == nil || deadline == 0 {
		return
	}
	go func() {
		time.Sleep(deadline)
		ctx.DeleteMessage(msg.Chat.ID, msg.MessageID)
	}()

}
