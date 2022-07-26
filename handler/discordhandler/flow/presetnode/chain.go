package presetnode

import (
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/bot-wizard/pconst"
)

func GetChainOption() *discordgo.ApplicationCommandOption {

	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0)
	for _, v := range pconst.ChainTypeList {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  pconst.GetChainName(uint32(v)),
			Value: v,
		})
	}
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        "chain_type",
		Description: "Select Chain",
		Required:    true,
		Choices:     choices,
	}
}
