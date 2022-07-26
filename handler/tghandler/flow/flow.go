package flow

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/bot-wizard/handler/tghandler/tcontext"
)

type TGFlowHandler interface {
	Handle(ctx *tcontext.Context) error
	GetCmdParser() func(u *tgbotapi.Update) string
}
