package presetnode

import (
	"github.com/bwmarrin/discordgo"
)

type OptionAddressPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
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
		option.Required = payload.Required
	}

	return option
}
