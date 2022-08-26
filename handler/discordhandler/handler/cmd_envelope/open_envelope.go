package cmd_envelope

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/mdparse"
	"github.com/tristan-club/wizard/pkg/tstore"
	"strconv"
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
	envelopeId, err := strconv.ParseInt(ctx.Cid.GetId(), 10, 64)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "invalid envelopeId", "error": err.Error(), "ctx": ctx}).Send()
		return he.NewServerError(he.ServerError, "", fmt.Errorf("invalid red envelope id payload %s", err.Error()))
	}
	channelId := ctx.IC.ChannelID
	//assetSymbol := pconst.GetAssetSymbol(payload.ChainType)

	openEnvelopeResp, err := ctx.CM.OpenEnvelope(ctx.Context, &controller_pb.OpenEnvelopeReq{
		Address:    ctx.Requester.RequesterDefaultAddress,
		EnvelopeNo: "",
		EnvelopeId: uint32(envelopeId),
		IsWait:     false,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	}

	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if openEnvelopeResp.CommonResponse.Code != he.Success {
		if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_OPENED || openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_SOLD_OUT {
			if _, err := ctx.FollowUpReply(fmt.Sprintf(mdparse.ParseV2(text.BusinessError), ctx.GetNickNameMDV2(), "open envelope command", mdparse.ParseV2(openEnvelopeResp.CommonResponse.Message))); err != nil {
				log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
				return he.NewServerError(he.CodeBotSendMsgError, "", err)
			}

			if openEnvelopeResp.CommonResponse.Code == pconst.CODE_ENVELOPE_SOLD_OUT {
				msgId, err := tstore.PBGetStr(fmt.Sprintf("%s%d", pconst.EnvelopeStorePrefix, envelopeId), pconst.EnvelopeStorePath)
				if err != nil {
					log.Error().Fields(map[string]interface{}{"action": "get envelope msg error", "error": err.Error(), "id": envelopeId}).Send()
					return nil
				}
				err = ctx.Session.ChannelMessageDelete(ctx.GetGroupChannelId(), msgId)
				if err != nil {
					log.Error().Fields(map[string]interface{}{"action": "delete red envelope error", "error": err.Error(), "id": envelopeId}).Send()
				}
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

	_, err = ctx.FollowUpReply(fmt.Sprintf(text.OpenEnvelopeTransactionProcessing, ctx.GetNickNameMDV2(), envelopeId, amountLabel, assetSymbol))
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}

	time.Sleep(time.Second * 5)

	getDataResp, err := ctx.CM.GetTx(context.Background(), &controller_pb.GetTxReq{TxId: openEnvelopeResp.Data.TxId, IsWait: true})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if getDataResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(getDataResp.CommonResponse)
	}

	log.Info().Msgf("user %s open envelope %d tx hash %s", ctx.GetFromId(), envelopeId, getDataResp.Data.TxHash)

	if _, err = ctx.Send(channelId, fmt.Sprintf(text.OpenEnvelopeSuccess, ctx.GetNickNameMDV2(), envelopeId, mdparse.ParseV2(amount),
		mdparse.ParseV2(assetSymbol), chain_info.GetExplorerTargetUrl(net.ChainId, getDataResp.Data.TxHash, chain_info.ExplorerTargetTransaction))); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "bot send msg", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeBotSendMsgError, "", err)
	}
	return nil
}
