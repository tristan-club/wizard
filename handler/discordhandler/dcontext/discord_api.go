package dcontext

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/wizard/handler/discordhandler/rate"
	"github.com/tristan-club/wizard/handler/text"
)

func (ctx *Context) DM(content string) error {

	rate.CheckLimit(ctx.GetChatId())

	channel, err := ctx.Session.UserChannelCreate(ctx.GetFromId())
	if err != nil {
		return err
	}

	_, err = ctx.Session.ChannelMessageSend(channel.ID, content)
	return err
}

func (ctx *Context) AckMsg(ephemeralMsgInPrivate bool) error {

	rate.CheckLimit(ctx.IC.Interaction.ChannelID)

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

func (ctx *Context) FollowUpReply(content string) (*discordgo.Message, error) {

	wp := &discordgo.WebhookParams{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        "rich",
			Description: content,
		},
	}}
	return ctx.FollowUpReplyComplex(wp)
}

func (ctx *Context) FollowUpReplyComplex(wp *discordgo.WebhookParams) (*discordgo.Message, error) {

	rate.CheckLimit(ctx.IC.Interaction.ChannelID)

	msg, err := ctx.Session.FollowupMessageCreate(ctx.IC.Interaction, true, wp)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (ctx *Context) FollowUpEdit(messageId string, content string) error {

	wp := &discordgo.WebhookEdit{
		Content: &content,
	}

	return ctx.FollowUpEditComplex(messageId, wp)
}

func (ctx *Context) FollowUpEditComplex(messageId string, wp *discordgo.WebhookEdit) error {

	rate.CheckLimit(ctx.IC.Interaction.ChannelID)

	_, err := ctx.Session.FollowupMessageEdit(ctx.IC.Interaction, messageId, wp)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *Context) Reply(content string, ephemeralMsg bool) error {

	rate.CheckLimit(ctx.IC.Interaction.ChannelID)

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

	we := &discordgo.WebhookEdit{
		Content:         &content,
		Components:      nil,
		Embeds:          nil,
		Files:           nil,
		AllowedMentions: nil,
	}

	return ctx.EditReplyComplex(we)
	//return msg, er
}

func (ctx *Context) EditReplyComplex(we *discordgo.WebhookEdit) (*discordgo.Message, error) {

	rate.CheckLimit(ctx.IC.Interaction.ChannelID)

	return ctx.Session.InteractionResponseEdit(ctx.IC.Interaction, we)
	//return msg, er
}

func (ctx *Context) ReplyDmWithGroupForward(groupChannelId, userId, content string) error {

	if userId == "" && ctx.IC.Member != nil {
		userId = ctx.IC.Member.User.ID
	}

	if groupChannelId == "" {
		groupChannelId = ctx.IC.ChannelID
	}

	rate.CheckLimit("")

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

	msg := &discordgo.MessageSend{Content: content}
	return ctx.SendComplex(chatId, msg)

}

func (ctx *Context) SendComplex(chatId string, message *discordgo.MessageSend) (*discordgo.Message, error) {

	rate.CheckLimit(chatId)

	return ctx.Session.ChannelMessageSendComplex(chatId, message)
}
