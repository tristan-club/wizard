package handler

import (
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
)

type DiscordCmdHandler struct {
	ApplicationCommand *discordgo.ApplicationCommand     `json:"detail"`
	Handler            func(ctx *dcontext.Context) error `json:"handler"`
}

func (d DiscordCmdHandler) Handle(ctx *tcontext.Context) error {
	//TODO implement me
	panic("implement me")
}

func (d DiscordCmdHandler) GetCmdParser() func(u *tgbotapi.Update) string {
	//TODO implement me
	panic("implement me")
}
