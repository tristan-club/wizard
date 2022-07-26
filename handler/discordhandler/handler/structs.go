package handler

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/dcontext"
)

type DiscordCmdHandler struct {
	ApplicationCommand *discordgo.ApplicationCommand     `json:"detail"`
	Handler            func(ctx *dcontext.Context) error `json:"handler"`
}
