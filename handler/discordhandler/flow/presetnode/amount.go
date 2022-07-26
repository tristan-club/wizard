package presetnode

import (
	"github.com/bwmarrin/discordgo"
)

func GetAmountOption() *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "amount",
		Description: "Enter Amount",
		Required:    true,
	}
}
