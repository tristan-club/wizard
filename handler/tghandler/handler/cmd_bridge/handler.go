package cmd_bridge

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/bot-wizard/cmd"
	"github.com/tristan-club/bot-wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/bot-wizard/handler/text"
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/bot-wizard/handler/tghandler/handler/cmd_bridge/bridge_node"
	"github.com/tristan-club/bot-wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/bot-wizard/handler/userstate"
	"github.com/tristan-club/bot-wizard/pconst"
	"github.com/tristan-club/bot-wizard/pkg/chain_info"
	he "github.com/tristan-club/bot-wizard/pkg/error"
	"github.com/tristan-club/bot-wizard/pkg/log"
	"io"
	"strings"
)

var fromChainType = uint32(chain_info.ChainTypeBsc)
var fromChainId = pconst.GetChainId(fromChainType)
var toChainType = uint32(chain_info.ChainTypeMetis)
var toChainId = pconst.GetChainId(toChainType)

type BridgePayload struct {
	UserNo      string `json:"user_no"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	ConfirmText string `json:"confirm_text"`
	PinCode     string `json:"pin_code"`
}

var Handler *chain.ChainHandler

func init() {

	Handler = chain.NewChainHandler(cmd.CmdBridge, bridgeSendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(bridge_node.EnterBridgeAmountNode, nil).
		AddPresetNode(presetnode.EnterPinCodeHandler, &presetnode.EnterPinCodeParam{
			Content:             "",
			ParamKey:            "",
			UseTargetContentKey: "bridge_content",
			UserMarkdown:        true,
		})
}

func bridgeSendHandler(ctx *tcontext.Context) error {

	var payload = &BridgePayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}

	bridgeReq := &controller_pb.BridgeReq{
		FromChainType: fromChainType,
		FromChainId:   fromChainId,
		ToChainType:   toChainType,
		ToChainId:     toChainId,
		BridgeAmount:  payload.Amount,
		FromId:        payload.UserNo,
		From:          payload.From,
		PinCode:       payload.PinCode,
	}

	bridgeResp, err := ctx.CM.Bridge(ctx.Context, bridgeReq)
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if bridgeResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(bridgeResp.CommonResponse)
	}

	brNo := bridgeResp.Data.RecordNo
	var thisMsg *tgbotapi.Message
	if thisMsg, herr = ctx.Send(ctx.U.SentFrom().ID, text.BridgeSubmitted, nil, true, false); herr != nil {
		return herr
	}

	requesterCtx, herr := ctx.CopyRequester()
	if herr != nil {
		return herr
	}
	brStream, err := ctx.CM.WaitBridgeData(requesterCtx, &controller_pb.GetBridgeReq{RecordNo: brNo})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "get bridge record stream", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if bridgeResp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "get bridge record stream", "error": bridgeResp}).Send()
		return tcontext.RespToError(bridgeResp.CommonResponse)
	}

	lastState := uint32(pconst.BridgeStatusCreate)
	//currentCheckTime := 0
	for {
		//if currentCheckTime > 20*30 {
		//	log.Error().Fields(map[string]interface{}{"action": "bridge out of time", "error": err.Error()}).Send()
		//	return he.NewError(he.CodeWalletRequestError, fmt.Errorf("transaction out of time"))
		//}
		//time.Sleep(time.Second * 3)
		//currentCheckTime++

		bridgeRecordResp, err := brStream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return he.NewServerError(he.CodeWalletRequestError, "", err)
			}
		} else if bridgeRecordResp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "get bridge record ", "error": bridgeResp}).Send()
			return tcontext.RespToError(bridgeResp.CommonResponse)
		}
		currentStatus := bridgeRecordResp.Data.Status
		if lastState == currentStatus {
			log.Debug().Fields(map[string]interface{}{"action": "get pending tx", "br no": brNo, "status": currentStatus}).Send()
			continue
		}
		respContent := ""
		switch currentStatus {
		case pconst.BridgeStatusFromTxPending:

			respContent = fmt.Sprintf(text.BridgeFromPending, fmt.Sprintf("%s%s", pconst.GetExplore(fromChainType, pconst.ExploreTypeTx), bridgeRecordResp.Data.FromTxHash))

			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, respContent, nil, true, false); herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send msg", "error": herr}).Send()
				return herr
			}

		case pconst.BridgeStatusFromTxSuccess:
			respContent = fmt.Sprintf(text.BridgeFromSuccess)

			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, respContent, nil, true, false); herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send msg", "error": herr}).Send()
				return herr
			}

		case pconst.BridgeStatusFromTxFailed:
			respContent = fmt.Sprintf(text.BridgeTransactionFailed, fmt.Sprintf("%s%s", pconst.GetExplore(fromChainType, pconst.ExploreTypeTx), bridgeRecordResp.Data.FromTxHash))
			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, respContent, nil, true, false); herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send msg", "error": herr}).Send()
				return herr
			}
			return nil

		case pconst.BridgeStatusToTxPending:

			respContent = fmt.Sprintf(text.BridgeToPending, fmt.Sprintf("%s%s", pconst.GetExplore(toChainType, pconst.ExploreTypeTx), bridgeRecordResp.Data.ToTxHash))

			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, respContent, nil, true, false); herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send msg", "error": herr}).Send()
				return herr
			}

		case pconst.BridgeStatusToTxSuccess:

			metisLatestBalance := "get latest balance error:"
			assetResp, err := ctx.CM.GetAsset(ctx.Context, &controller_pb.GetAssetReq{
				ChainType:       toChainType,
				ChainId:         toChainId,
				Address:         ctx.Requester.RequesterDefaultAddress,
				ContractAddress: "",
				ForceBalance:    true,
			})
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "get balance", "error": err.Error()}).Send()
				metisLatestBalance = fmt.Sprintf("%s%s", metisLatestBalance, err.Error())
			} else if assetResp.CommonResponse.Code != he.Success {
				log.Error().Fields(map[string]interface{}{"action": "get balance", "error": assetResp}).Send()
				metisLatestBalance = fmt.Sprintf("%s%s", metisLatestBalance, assetResp.CommonResponse)
			} else {
				metisLatestBalance = assetResp.Data.BalanceCutDecimal
			}

			respContent = fmt.Sprintf(text.BridgeToSuccess, payload.Amount, metisLatestBalance, fmt.Sprintf("%s%s", pconst.GetExplore(toChainType, pconst.ExploreTypeTx), bridgeRecordResp.Data.ToTxHash))
			if strings.Contains(respContent, ".") {
				respContent = strings.ReplaceAll(respContent, ".", "\\.")
			}
			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, respContent, nil, true, false); herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send msg", "error": herr}).Send()
				return herr
			}
			return nil

		case pconst.BridgeStatusToTxFailed:
			respContent = fmt.Sprintf(text.BridgeTransactionFailed, fmt.Sprintf("%s%s", pconst.GetExplore(toChainType, pconst.ExploreTypeTx), bridgeRecordResp.Data.FromTxHash))
			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, respContent, nil, true, false); herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send msg", "error": herr}).Send()
				return herr
			}
			return nil
		case pconst.BridgeStatusOtherFailed:
			respContent = fmt.Sprintf(text.BridgeOtherFailed, bridgeRecordResp.Data.ErrMsg)
			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, respContent, nil, true, false); herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send msg", "error": herr}).Send()
				return herr
			}
			return nil
		}

		lastState = currentStatus
	}
}
