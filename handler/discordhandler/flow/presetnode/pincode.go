package presetnode

import "github.com/bwmarrin/discordgo"

func GetPinCodeOption(name, desc string) *discordgo.ApplicationCommandOption {
	if desc == "" {
		desc = "Enter your pin code"
	}
	if name == "" {
		name = "pin_code"
	}
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: desc,
		Required:    true,
	}
}
