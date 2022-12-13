package cmd_envelope

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/bignum"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/mdparse"
	"github.com/tristan-club/kit/tstore"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/util"
	"math/big"
	"strings"
	"time"
)

type OpenEnvelopePayload struct {
	EnvelopeId uint32 `json:"envelope_no"`
}

var OpenEnvelopeHandler = &handler.DiscordCmdHandler{

	ApplicationCommand: &discordgo.ApplicationCommand{
		ApplicationID: "",
		DMPermission:  &dmPermissionFalse,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "envelope_no",
				Description: "Enter red envelope NO",
				Required:    true,
			},
		},
		Version: "1",
	},
	Handler: openEnvelopeHandler,
}

func openEnvelopeHandler(ctx *dcontext.Context) error {

	//wp2 := &discordgo.MessageSend{
	//	Embeds: []*discordgo.MessageEmbed{
	//		{
	//			Type:  "rich",
	//			Title: "",
	//			Description: "[click](https://baidu.com)\n" +
	//				"MSG: [URL](https://discord.com/channels/991610208096882788/1006408396934746162/1047444181783674940)",
	//		},
	//	},
	//	Reference: nil,
	//}
	//
	//if _, err := ctx.SendComplex(ctx.IC.ChannelID, wp2); err != nil {
	//	log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
	//	return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	//}
	//return nil

	envelopeNo := ctx.Cid.GetId()
	channelId := ctx.IC.ChannelID
	//assetSymbol := pconst.GetAssetSymbol(payload.ChainType)

	param := &StartParam{}
	if !util.IsNil(ctx.Param) {
		if _param, ok := ctx.Param.(*StartParam); ok {
			log.Info().Fields(map[string]interface{}{"action": "get start param", "param": ctx.Param}).Send()
			param = _param
		}
	}

	openEnvelopeResp, err := ctx.CM.OpenEnvelope(ctx.Context, &controller_pb.OpenEnvelopeReq{
		Address:        ctx.Requester.RequesterDefaultAddress,
		EnvelopeNo:     envelopeNo,
		IsWait:         false,
		ReceiverNo:     ctx.Requester.RequesterUserNo,
		EnvelopeOption: controller_pb.ENVELOPE_OPTION(ctx.Cid.GetCallbackType()),
	})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	}

	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if openEnvelopeResp.CommonResponse.Code != he.Success {
		if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_OPENED || openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_SOLD_OUT {
			if _, err := ctx.FollowUpReply(fmt.Sprintf(mdparse.ParseV2(text.BusinessError), ctx.GetNickNameMDV2(), "open envelope command", mdparse.ParseV2(openEnvelopeResp.CommonResponse.Message))); err != nil {
				log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
				return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
			}

			return nil
		} else if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOP_STATE_INVALID {
			// todo
			return tcontext.RespToError(openEnvelopeResp.CommonResponse)
		} else {
			return tcontext.RespToError(openEnvelopeResp.CommonResponse)
		}

	}

	amount := openEnvelopeResp.Data.Amount
	assetSymbol := openEnvelopeResp.Data.AssetSymbol
	chainType := openEnvelopeResp.Data.ChainType
	net := chain_info.GetNetByChainType(chainType)
	amountLabel := strings.ReplaceAll(amount, ".", "\\.")

	_, err = ctx.FollowUpReply(fmt.Sprintf(text.OpenEnvelopeTransactionProcessing, ctx.GetNickNameMDV2(), envelopeNo, amountLabel, assetSymbol))
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}

	time.Sleep(time.Second * 5)

	getDataResp, err := ctx.CM.GetTx(context.Background(), &controller_pb.GetTxReq{TxId: openEnvelopeResp.Data.TxId, IsWait: true})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "request controller svc error", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if getDataResp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "get envelope tx error", "error": getDataResp}).Send()
		return tcontext.RespToError(getDataResp.CommonResponse)
	}

	log.Info().Msgf("user %s open envelope %s tx hash %s", ctx.GetFromId(), envelopeNo, getDataResp.Data.TxHash)

	envelopeMsgId, err := tstore.PBGetStr(fmt.Sprintf("%s%s", pconst.EnvelopeStorePrefix, envelopeNo), pconst.EnvelopeStorePathMsgId)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "get envelope msg id error", "error": err.Error()}).Send()
	} else {

		//var title string
		//
		//title, err = tstore.PBGetStr(fmt.Sprintf("%s%s", pconst.EnvelopeStorePrefix, envelopeNo), pconst.EnvelopeStorePathTitle)
		//if err != nil {
		//	log.Error().Fields(map[string]interface{}{"action": "get title error", "error": err.Error(), "envelopeNo": envelopeNo}).Send()
		//}

		getEnvelopeResp, err := ctx.CM.GetEnvelope(ctx.Context, &controller_pb.GetEnvelopeReq{EnvelopeNo: envelopeNo, WithClaimList: true, WaitSuccess: true})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "call wallet", "error": err.Error()}).Send()
		} else if getEnvelopeResp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "get envelope", "error": getEnvelopeResp.CommonResponse}).Send()
		} else {
			title := text.EnvelopeTitleOrdinary
			if controller_pb.ENVELOPE_OPTION(ctx.Cid.GetCallbackType()) == controller_pb.ENVELOPE_OPTION_HAS_CAT {
				title = text.EnvelopeTitleCAT
			}
			title = fmt.Sprintf(title, envelopeNo, getEnvelopeResp.Data.CreatorName)

			envelopeDetail := fmt.Sprintf(text.EnvelopeDetail,
				mdparse.ParseV2(bignum.CutDecimal(new(big.Int).SetUint64(getEnvelopeResp.Data.RemainAmount), 4, 4)), mdparse.ParseV2(getEnvelopeResp.Data.AssetSymbol),
				getEnvelopeResp.Data.Quantity-getEnvelopeResp.Data.RemainQuantity, getEnvelopeResp.Data.Quantity)

			var claimHistory string
			for _, claim := range getEnvelopeResp.Data.ClaimList {
				labelLen := 20
				labelName := make([]rune, labelLen)
				nicknameRune := []rune(claim.ReceiverNickname)
				for i := 0; i < len(nicknameRune) && i < labelLen; i++ {
					labelName[i] = nicknameRune[i]
				}

				// todo 没有txhash，用address代替
				var txnUrl string
				if claim.TxHash != "" {
					txnUrl = chain_info.GetExplorerTargetUrl(getEnvelopeResp.Data.ChainId, claim.TxHash, chain_info.ExplorerTargetTransaction)
				} else {
					txnUrl = chain_info.GetExplorerTargetUrl(getEnvelopeResp.Data.ChainId, claim.ReceiverAddress, chain_info.ExplorerTargetAddress)
				}

				claimHistory += fmt.Sprintf("%s   %s %s   TXN URL\\: [click to view](%s)\n\n",
					ctx.GenerateNickName(claim.ReceiverOpenId),
					mdparse.ParseV2(claim.Amount),
					mdparse.ParseV2(claim.AssetSymbol), mdparse.ParseV2(txnUrl))
			}
			envelopeDetail = envelopeDetail + "\n\n🎊Claim History:\n\n" + claimHistory

			var openButton *tgbotapi.InlineKeyboardMarkup
			if getEnvelopeResp.Data.RemainQuantity != 0 {
				openButton = &tgbotapi.InlineKeyboardMarkup{}
				*openButton = tgbotapi.NewInlineKeyboardMarkup(
					[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text.OpenEnvelope, ctx.Cid.String())},
				)
			}

			e := &discordgo.MessageEmbed{
				Type:        "rich",
				Title:       title,
				Description: envelopeDetail,
			}

			if param.Photo != "" {
				e.Image = &discordgo.MessageEmbedImage{URL: param.Photo}
			}

			me := &discordgo.MessageEdit{
				Channel: channelId,
				ID:      envelopeMsgId,
				Embeds: []*discordgo.MessageEmbed{
					e,
				},
			}

			cid := ctx.Cid
			if param.CustomType != 0 {
				cid = customid.NewCustomId(param.CustomType, envelopeNo, param.Option)
			}

			bt := &discordgo.Button{
				Label:    text.OpenEnvelope,
				Style:    discordgo.PrimaryButton,
				Disabled: false,
				CustomID: cid.String(),
			}

			if getEnvelopeResp.Data.RemainAmount == 0 {
				bt.Disabled = true
			}
			me.Components = []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						bt,
					},
				},
			}

			_, err = ctx.RatedSession().ChannelMessageEditComplex(me)
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "edit envelope error", "error": err.Error(), "wp": me, "msgId": envelopeMsgId}).Send()
			}

		}
	}

	wp := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Type:  "rich",
				Title: "",
				Description: fmt.Sprintf(text.OpenEnvelopeSuccessGroupMsg, ctx.GetNickNameMDV2(), envelopeNo, mdparse.ParseV2(amount),
					mdparse.ParseV2(assetSymbol), chain_info.GetExplorerTargetUrl(net.ChainId, getDataResp.Data.TxHash, chain_info.ExplorerTargetTransaction),
					pconst.DCLink+fmt.Sprintf("/%s/%s/%s", ctx.IC.GuildID, ctx.IC.ChannelID, envelopeMsgId)),
			},
		},
		Reference: nil,
	}

	if _, err = ctx.SendComplex(channelId, wp); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
	}
	return nil
}
