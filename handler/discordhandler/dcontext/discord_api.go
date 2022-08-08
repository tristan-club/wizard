package dcontext

import (
	"github.com/bwmarrin/discordgo"
)

func (ctx *Context) DM(content string) error {

	channel, err := ctx.Session.UserChannelCreate(ctx.GetFromId())
	if err != nil {
		return err
	}

	_, err = ctx.Session.ChannelMessageSend(channel.ID, content)
	return err
}

func (ctx *Context) Reply(content string, ephemeralMsg bool) error {

	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	}

	if !ctx.IsPrivate() || ephemeralMsg {
		resp.Data.Flags = discordgo.MessageFlagsEphemeral
	}

	err := ctx.Session.InteractionRespond(ctx.IC.Interaction, resp)
	if err != nil {
		return err
	}

	ctx.IsICResponded = true

	return nil
}

func (ctx *Context) EditReply(content string) (*discordgo.Message, error) {
	return ctx.Session.InteractionResponseEdit(ctx.IC.Interaction, &discordgo.WebhookEdit{
		Content:         &content,
		Components:      nil,
		Embeds:          nil,
		Files:           nil,
		AllowedMentions: nil,
	})
	//return msg, er
}

func (ctx *Context) ReplyDmWithGroupForward(groupChannelId, userId, content string) error {

	if userId == "" && ctx.IC.Member != nil {
		userId = ctx.IC.Member.User.ID
	}

	if groupChannelId == "" {
		groupChannelId = ctx.IC.ChannelID
	}

	err := ctx.Session.InteractionRespond(ctx.IC.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Please forward to DM with bot for detail",
		},
	})
	if err != nil {
		return err
	}

	if content != "" {
		channel, err := ctx.Session.UserChannelCreate(userId)
		if err != nil {
			return err
		}

		_, err = ctx.Session.ChannelMessageSend(channel.ID, content)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ctx *Context) Send(chatId string, content string) (*discordgo.Message, error) {

	if chatId == "" {
		chatId = ctx.IC.ChannelID
	}
	return ctx.Session.ChannelMessageSend(chatId, content)

}
