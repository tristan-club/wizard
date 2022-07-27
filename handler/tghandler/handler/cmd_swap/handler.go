package cmd_swap

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/prehandler"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmd_swap/swap_node"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/bignum"
	"github.com/tristan-club/wizard/pkg/dingding"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/log"
	"io"
	"math/big"
	"os"
	"strings"
	"time"
)

type SwapPayload struct {
	UserNo       string `json:"user_no"`
	ChainType    uint32 `json:"chain_type"`
	From         string `json:"from"`
	ChannelId    string `json:"channel_id"`
	OriginToken  string `json:"origin_token"`
	Amount       string `json:"amount"`
	TargetToken  string `json:"target_token"`
	TargetAmount string `json:"target_amount"`
	SwapRate     string `json:"swap_rate"`
	TokenType    uint32 `json:"token_type"`
	Gas          string `json:"gas"`
	PinCode      string `json:"pin_code"`
}

var fromChainType = uint32(chain_info.ChainTypeBsc)

const (
	maxQueryTime = 20 * 30
)

var Handler *chain.ChainHandler

//Todo optimize code
func init() {
	var swapAssetList []string
	var swapAssetValueList []int64
	for k, v := range pconst.GetSwapAssetList(chain_info.ChainTypeBsc) {
		swapAssetList = append(swapAssetList, v.AssetSymbol)
		swapAssetValueList = append(swapAssetValueList, int64(k))
	}

	Handler = chain.NewChainHandler(cmd.CmdSwap, swapAndBridgeSendHandler).
		AddPreHandler(prehandler.ForwardPrivate).
		AddPreHandler(prehandler.SetFrom).
		AddPresetNode(presetnode.EnterTypeNode, &presetnode.EnterTypeParam{
			ChoiceText:         swapAssetList,
			ChoiceValue:        swapAssetValueList,
			Content:            text.EnterBridgeAsset,
			ParamKey:           "",
			ChosenTextParamKey: "asset_symbol",
		}).
		AddPresetNode(swap_node.EnterSwapAndBridgeAmountNode, nil).
		AddPresetNode(presetnode.EnterPinCodeHandler, &presetnode.EnterPinCodeParam{
			Content:             "",
			ParamKey:            "",
			UseTargetContentKey: "swap_content",
			UserMarkdown:        true,
		})
}

func swapAndBridgeSendHandler(ctx *tcontext.Context) error {

	var payload = &SwapPayload{}
	_, herr := userstate.GetState(ctx.OpenId(), payload)
	if herr != nil {
		return herr
	}
	swapReq := &controller_pb.SwapReq{
		ChainType:          payload.ChainType,
		ChainId:            pconst.GetChainId(payload.ChainType),
		FromId:             payload.UserNo,
		From:               payload.From,
		TokenType:          payload.TokenType,
		OriginTokenAddress: payload.OriginToken,
		Value:              payload.Amount,
		PinCode:            payload.PinCode,
		CheckBalance:       true,
		TargetTokenAddress: payload.TargetToken,
		TargetValue:        payload.TargetAmount,
		ChannelId:          payload.ChannelId,
		IsWait:             true,
	}
	var thisMsg *tgbotapi.Message
	shouldApprove := false
	swapReq.ContractAddress = pconst.PANCAKE_BSC
	swapReq.To = swapReq.From
	if swapReq.TokenType == pconst.TokenTypeInternal {
		log.Info().Msgf("[swap] don`t need approve")
		swapReq.OriginTokenAddress = pconst.TOKEN_BSC_WBNB
	} else {
		//check allowance
		log.Info().Msgf("[swap] check allowance")
		allowanceResp, err := ctx.CM.Allowance(ctx.Context, &controller_pb.AllowanceReq{
			Owner:     swapReq.From,
			Spender:   swapReq.ContractAddress,
			ChainId:   swapReq.ChainId,
			ChainType: swapReq.ChainType,
			TokenAddr: swapReq.OriginTokenAddress,
		})

		if err != nil {
			return he.NewServerError(he.CodeWalletRequestError, "", err)
		} else if allowanceResp.Code != he.Success {
			return tcontext.RespToError(allowanceResp.Message)
		}

		allowanceV, ok1 := new(big.Int).SetString(allowanceResp.AllowanceAmount, 10)
		needV, ok2 := bignum.HandleAddDecimal(swapReq.Value, 18)

		if !ok1 || !ok2 {
			log.Error().Msgf("str to big int error:allowanceV:%s,needV:%s", allowanceV, needV)
			return he.NewServerError(he.CodeInvalidPayload, "", fmt.Errorf("str to big int error:allowanceV:%s,needV:%s", allowanceV, needV))
		}

		shouldApprove = allowanceV.Cmp(needV) < 0
		log.Info().Msgf("[swap] need approve:%v,allowanceV:%s,needV %s", shouldApprove, allowanceV.String(), needV.String())
		//if need approve
		if shouldApprove {
			log.Info().Msgf("[swap] do allowance")
			if thisMsg, herr = ctx.Send(ctx.U.SentFrom().ID, text.SwapApproveStart, nil, true, false); herr != nil {
				return herr
			}
			approveResp, err := ctx.CM.Approve(ctx.Context, &controller_pb.ApproveReq{
				From:      swapReq.From,
				Spender:   swapReq.ContractAddress,
				TokenAddr: swapReq.OriginTokenAddress,
				PinCode:   swapReq.PinCode,
				FromId:    swapReq.FromId,
				ChainId:   swapReq.ChainId,
				ChainType: swapReq.ChainType,
				Value:     needV.Mul(needV, big.NewInt(10)).String(),
				IsWait:    false,
			})
			if err != nil {
				return he.NewServerError(he.CodeWalletRequestError, "", err)
			} else if approveResp.CommonResponse.Code != he.Success {
				return tcontext.RespToError(approveResp.CommonResponse.Message)
			}
			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, fmt.Sprintf(text.SwapApproveProcessing, fmt.Sprintf("%s%s", pconst.GetExplore(payload.ChainType, pconst.ExploreTypeTx), approveResp.Data.TxHash)), nil, true, false); herr != nil {
				return herr
			}
			txHashResp, err := ctx.CM.GetTx(ctx.Context, &controller_pb.GetTxReq{
				TxHash: approveResp.Data.TxHash,
				IsWait: true,
			})
			if err != nil {
				return he.NewServerError(he.CodeWalletRequestError, "", err)
			} else if approveResp.CommonResponse.Code != he.Success {
				return tcontext.RespToError(approveResp.CommonResponse.Message)
			}

			log.Info().Msgf("[swap] get approve tx result:%v", txHashResp)
			if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, text.SwapApproveToSwap, nil, true, false); herr != nil {
				return herr
			}
		}
	}

	//swap

	swapResp, err := ctx.CM.Swap(ctx.Context, swapReq)
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if swapResp.CommonResponse.Code != he.Success {
		return tcontext.RespToError(swapResp.CommonResponse.Message)
	}

	var swapContent string
	if shouldApprove {
		swapContent = fmt.Sprintf(text.SwapProcessing1, fmt.Sprintf("%s%s", pconst.GetExplore(payload.ChainType, pconst.ExploreTypeTx), swapResp.Data.TxHash))

	} else {
		swapContent = fmt.Sprintf(text.SwapProcessing2, fmt.Sprintf("%s%s", pconst.GetExplore(payload.ChainType, pconst.ExploreTypeTx), swapResp.Data.TxHash))
	}
	if thisMsg != nil {
		if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, swapContent, nil, true, false); herr != nil {
			return herr
		}
	} else {
		if thisMsg, herr = ctx.Send(ctx.U.SentFrom().ID, swapContent, nil, true, false); herr != nil {
			return herr
		}
	}
	requesterCtx, herr := ctx.CopyRequester()
	if herr != nil {
		return herr
	}

	swapStream, err := ctx.CM.WaitSwapData(requesterCtx, &controller_pb.GetSwapReq{RecordNo: swapResp.Data.RecordNo})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "get swap record stream", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if swapResp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "get swap record stream", "error": swapResp}).Send()
		return tcontext.RespToError(swapResp.CommonResponse)
	}

	lastState := swapResp.Data.Status
	var endContent string
	success := true

	//wait for swapping result
	//try max time 10
	//try wait time 1s
	currentCheckTime := 0
	for {
		if currentCheckTime >= maxQueryTime {
			hashUri := fmt.Sprintf("%s%s", pconst.GetExplore(fromChainType, pconst.ExploreTypeTx))
			if shouldApprove {
				endContent = fmt.Sprintf(text.CheckTimeOut1, fmt.Sprintf("%s%s", hashUri, swapResp.Data.TxHash))
			} else {
				endContent = fmt.Sprintf(fmt.Sprintf(text.CheckTimeout2, hashUri, swapResp.Data.TxHash))
			}
			success = false
			dingDingToken := os.Getenv("DING_TOKEN")
			dingDingSecret := os.Getenv("DING_SECRET")
			if dingDingToken != "" {
				bot := dingding.NewRobot(dingDingToken, dingDingSecret, "", "")
				if err := bot.SendMarkdownMessage("## MetaWallet Swap Warn", fmt.Sprintf("This swap transaction proccessing time out\nuri:%s\nhash:%s\nswap record no:%s\nplease check it", hashUri, swapResp.Data.TxHash, swapResp.Data.RecordNo), []string{}, false); err != nil {
					log.Error().Fields(map[string]interface{}{"action": "ding ding send", "error": err.Error()}).Send()
				}
			}
			break
		}
		currentCheckTime++
		time.Sleep(3 * time.Second)

		swapRecResp, err := swapStream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				return he.NewServerError(he.CodeWalletRequestError, "", err)
			}
		} else if swapRecResp.CommonResponse.Code != he.Success {
			log.Error().Fields(map[string]interface{}{"action": "get swap record ", "error": swapRecResp}).Send()
			return tcontext.RespToError(swapRecResp.CommonResponse)
		}
		currentStatus := swapRecResp.Data.Status
		if lastState == currentStatus {
			log.Debug().Fields(map[string]interface{}{"action": "get pending tx", "br no": swapResp.Data.RecordNo, "status": currentStatus}).Send()
			continue
		}

		//generate content by tx status
		switch currentStatus {
		case pconst.TxStateFailed:
			success = false
			endContent = fmt.Sprintf(text.SwapFailed, swapRecResp.Data.ErrMsg,
				fmt.Sprintf("%s%s", pconst.GetExplore(fromChainType, pconst.ExploreTypeTx), swapRecResp.Data.TxHash))
		case pconst.TxStateSuccess:
			assetResp, err := ctx.CM.GetAsset(ctx.Context, &controller_pb.GetAssetReq{
				ChainType:       swapReq.ChainType,
				ChainId:         swapReq.ChainId,
				Address:         ctx.Requester.RequesterDefaultAddress,
				ContractAddress: swapReq.TargetTokenAddress,
				ForceBalance:    true,
			})
			if err != nil {
				return he.NewServerError(he.CodeWalletRequestError, "", err)
			} else if assetResp.CommonResponse.Code != he.Success {
				return tcontext.RespToError(assetResp.CommonResponse)
			}
			endContent = fmt.Sprintf(text.SwapSuccess, swapReq.TargetValue[0:8], assetResp.Data.BalanceCutDecimal, fmt.Sprintf("%s%s", pconst.GetExplore(fromChainType, pconst.ExploreTypeTx), swapRecResp.Data.TxHash))
		default:

		}
		lastState = currentStatus

		if len(endContent) > 0 {
			break
		}
	}
	var continueButton *tgbotapi.InlineKeyboardMarkup
	if success {
		openButton := tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(text.ContinueBridge, cmd.CmdBridge)},
		)
		continueButton = &openButton
	}

	endContent = strings.ReplaceAll(endContent, ".", "\\.")

	if herr = ctx.EditMessageAndKeyboard(ctx.U.SentFrom().ID, thisMsg.MessageID, endContent, continueButton, true, true); herr != nil {
		return herr
	}

	//ctx.SetDeadlineMsg(ctx.U.SentFrom().ID, thisMsg.MessageID, constant.COMMON_KEYBOARD_DEADLINE)

	return nil
}
