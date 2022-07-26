package presetnode

import (
	"github.com/bwmarrin/discordgo"
)

type OptionAddressPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func GetAddressOption(payload *OptionAddressPayload) *discordgo.ApplicationCommandOption {

	option := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "address",
		Description: "Enter Address",
		Required:    true,
	}

	if payload != nil {
		if payload.Name != "" {
			option.Name = payload.Name
		}
		if payload.Description != "" {
			option.Description = payload.Description
		}
	}

	return option
}
