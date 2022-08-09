package dcontext

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/wizard/handler/text"
)

func (ctx *Context) DM(content string) error {

	channel, err := ctx.Session.UserChannelCreate(ctx.GetFromId())
	if err != nil {
		return err
	}

	_, err = ctx.Session.ChannelMessageSend(channel.ID, content)
	return err
}

func (ctx *Context) AckMsg(ephemeralMsgInPrivate bool) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text.OperationProcessing,
		},
	}

	if !ctx.IsPrivate() || ephemeralMsgInPrivate {
		resp.Data.Flags = discordgo.MessageFlagsEphemeral
	}

	err := ctx.Session.InteractionRespond(ctx.IC.Interaction, resp)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *Context) FollowUpReply(content string) error {

	resp := &discordgo.WebhookParams{
		Content: content,
	}

	_, err := ctx.Session.FollowupMessageCreate(ctx.IC.Interaction, false, resp)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *Context) Reply(content string, ephemeralMsg bool) error {

	var icRespType uint8
	if icRespType == 0 {
		icRespType = uint8(discordgo.InteractionResponseChannelMessageWithSource)
	}

	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseType(icRespType),
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
