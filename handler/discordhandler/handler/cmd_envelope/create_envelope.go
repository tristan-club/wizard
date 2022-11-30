package cmd_envelope

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/mdparse"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/flow/presetnode"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/discordhandler/parser"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/tstore"
	"github.com/tristan-club/wizard/pkg/util"
	"strings"
)

type CreateEnvelopePayload struct {
	UserNo             string `json:"user_no"`
	From               string `json:"from"`
	ChainType          uint32 `json:"chain_type"`
	Asset              string `json:"token_address"`
	AssetSymbol        string `json:"asset_symbol"`
	EnvelopeType       uint32 `json:"envelope_type"`
	EnvelopeRewardType uint32 `json:"envelope_reward_type"`
	ChannelId          string `json:"channel_id"`
	Quantity           uint64 `json:"quantity"`
	Amount             string `json:"amount"`
	EnvelopeNo         string `json:"envelope_no"`
	EnvelopeOption     uint32 `json:"envelope_option"`
	PinCode            string `json:"pin_code"`
}

var dmPermissionFalse = false

var CreateEnvelopeHandler = &handler.DiscordCmdHandler{

	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		DMPermission:  &dmPermissionFalse,
		Options: []*discordgo.ApplicationCommandOption{
			presetnode.GetChainOption(),
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "envelope_reward_type",
				Description: "Select red envelope type",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Average Amount",
						Value: 1,
					},
					{
						Name:  "Random Amount",
						Value: 2,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "amount",
				Description: "Enter an amount between 0.0001 and 10000000000",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "quantity",
				Description: "Enter red envelope quantity. min 1, max 20",
				Required:    true,
			},

			presetnode.GetPinCodeOption("", ""),
			//presetnode.GetAddressOption(&presetnode.OptionAddressPayload{
			//	Name:        "token_address",
			//	Description: "If you enter this option, it will be regarded as using an added ERC20 token to send the red envelope",
			//	Required:    false,
			//}),
		},
		Version: "1",
	},
	Handler: envelopeSendHandler,
}

func envelopeSendHandler(ctx *dcontext.Context) error {

	var payload = &CreateEnvelopePayload{}

	err := parser.ParseOption(ctx.IC.Interaction, payload)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "parse param", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeInvalidPayload, "", err)
	}

	net := chain_info.GetNetByChainType(payload.ChainType)

	tokenType := pconst.TokenTypeInternal
	if payload.Asset != "" && payload.Asset != "INTERNAL" && strings.HasPrefix(payload.Asset, "0x") {
		addressChecked, err := util.EIP55Checksum(payload.Asset)
		if err != nil {
			log.Info().Fields(map[string]interface{}{"action": "address param invalid", "ctx": ctx}).Send()
			return he.NewServerError(pconst.CodeAddressParamInvalid, "", err)
		}
		payload.Asset = addressChecked
		tokenType = pconst.TokenTypeErc20

		tokenInfoResp, err := ctx.CM.GetToken(ctx.Context, &controller_pb.GetTokenInfoRequest{
			ChainId:         net.ChainId,
			ContractAddress: payload.Asset,
			TokenType:       uint32(tokenType),
		})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "request tokenInfo error", "error": err.Error()}).Send()
			return he.NewServerError(he.ServerError, "", err)
		} else if tokenInfoResp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "request tokenInfo error", "error": tokenInfoResp}).Send()
			return he.NewServerError(he.ServerError, tokenInfoResp.CommonResponse.Message, fmt.Errorf(tokenInfoResp.CommonResponse.Inner))
		}
		payload.AssetSymbol = tokenInfoResp.Data.Symbol

	} else {
		payload.AssetSymbol = net.Symbol
	}

	createEnvelopeReq := &controller_pb.AddEnvelopeReq{
		FromId:             ctx.Requester.RequesterUserNo,
		ChainType:          payload.ChainType,
		EnvelopeNo:         payload.EnvelopeNo,
		ChannelId:          ctx.GetGroupChannelId(),
		ChainId:            net.ChainId,
		TokenType:          uint32(tokenType),
		Address:            ctx.Requester.RequesterDefaultAddress,
		ContractAddress:    payload.Asset,
		Amount:             payload.Amount,
		Quantity:           payload.Quantity,
		EnvelopeType:       payload.EnvelopeType,
		EnvelopeRewardType: payload.EnvelopeRewardType,
		EnvelopeOption:     controller_pb.ENVELOPE_OPTION(payload.EnvelopeOption),
		Blessing:           "",
		PinCode:            payload.PinCode,
		IsWait:             false,
	}

	createRedEnvelope, err := ctx.CM.AddEnvelope(ctx.Context, createEnvelopeReq)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "add envelope error", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if createRedEnvelope.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "add envelope error", "error": createRedEnvelope}).Send()
		return tcontext.RespToError(createRedEnvelope.CommonResponse)
	}

	msg, err := ctx.FollowUpReply(fmt.Sprintf(text.EnvelopePreparing, pconst.GetExplore(payload.ChainType, createRedEnvelope.Data.AccountAddress, chain_info.ExplorerTargetAddress)))
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	requesterCtx, cancel, herr := ctx.CopyRequester()
	defer cancel()
	if herr != nil {
		return herr
	}

	//time.Sleep(time.Second * 1)
	envelopeResp, err := ctx.CM.GetEnvelope(requesterCtx, &controller_pb.GetEnvelopeReq{EnvelopeNo: createRedEnvelope.Data.EnvelopeNo, WaitSuccess: true})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "call wallet", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if envelopeResp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "get envelope", "error": envelopeResp.CommonResponse}).Send()
		return tcontext.RespToError(envelopeResp.CommonResponse)
	}

	if envelopeResp.Data.Status != pconst.EnvelopStatusRechargeSuccess {
		log.Error().Fields(map[string]interface{}{"action": "create envelope failed"}).Send()
		return he.NewBusinessError(0, text.EnvelopeCreateFailed, nil)
	}

	log.Debug().Fields(map[string]interface{}{"action": "create envelope success", "envelopeResp": envelopeResp})

	if err = ctx.FollowUpEdit(msg.ID, fmt.Sprintf(text.CreateEnvelopeSuccess, createRedEnvelope.Data.EnvelopeNo, chain_info.GetExplorerTargetUrl(net.ChainId, createRedEnvelope.Data.TxHash, chain_info.ExplorerTargetTransaction))); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	title := text.EnvelopeTitleOrdinary
	if payload.EnvelopeOption == uint32(controller_pb.ENVELOPE_OPTION_HAS_CAT) {
		title = text.EnvelopeTitleCAT
	}

	shareEnvelopeContent := fmt.Sprintf(text.EnvelopeDetailWithoutTitle, mdparse.ParseV2(payload.Amount), mdparse.ParseV2(payload.AssetSymbol), 0, payload.Quantity)

	cid := customid.NewCustomId(pconst.CustomIdOpenEnvelope, createRedEnvelope.Data.EnvelopeNo, int32(payload.EnvelopeOption))

	messageSend := &discordgo.MessageSend{
		Content: "",
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       title,
				Description: shareEnvelopeContent,
			},
		},
		TTS: false,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						CustomID: cid.String(),
						Disabled: false,
						Style:    discordgo.PrimaryButton,
						Label:    text.OpenEnvelope,
					},
				},
			},
		},
		Files:           nil,
		AllowedMentions: nil,
		Reference:       nil,
		File:            nil,
		Embed:           nil,
	}

	msg, err = ctx.SendComplex(ctx.GetGroupChannelId(), messageSend)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "send envelope msg error", "error": err.Error(), "ms": messageSend, "ctx": ctx}).Send()
	} else {
		err = tstore.PBSaveString(fmt.Sprintf("%s%s", pconst.EnvelopeStorePrefix, createRedEnvelope.Data.EnvelopeNo), pconst.EnvelopeStorePathMsgId, msg.ID)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "TStore save envelope message error", "error": err.Error()}).Send()
		}

		// todo 弄到 channel name
		err = tstore.PBSaveString(fmt.Sprintf("%s%s", pconst.EnvelopeStorePrefix, createRedEnvelope.Data.EnvelopeNo), pconst.EnvelopeStorePathChannelId, fmt.Sprintf("%s/%s", payload.ChannelId, ctx.IC.ChannelID))
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "TStore save envelope channel id error", "error": err.Error()}).Send()
		}
	}

	return nil
}
