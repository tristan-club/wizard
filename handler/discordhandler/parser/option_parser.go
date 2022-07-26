package parser

import (
	"encoding/json"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/bot-wizard/pkg/util"
)

func ParseOption(ic *discordgo.Interaction, input interface{}) error {
	if len(ic.ApplicationCommandData().Options) == 0 {
		return nil
	}

	if util.IsNil(input) {
		return nil
	}

	optionMap := make(map[string]interface{}, 0)

	for _, v := range ic.ApplicationCommandData().Options {
		optionMap[v.Name] = v.Value
	}

	b, err := json.Marshal(optionMap)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, &input); err != nil {
		return err
	}

	return nil

}

func OptionGetInt(ic *discordgo.Interaction) (int64, error) {
	if len(ic.ApplicationCommandData().Options) == 0 {
		return 0, errors.New("empty application command param")
	}

	return ic.ApplicationCommandData().Options[0].IntValue(), nil
}
