package tgmgr

import (
	"context"
	"encoding/base64"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow"
	"github.com/tristan-club/wizard/handler/tghandler/handler/cmdhandler"
	"github.com/tristan-club/wizard/handler/tghandler/inline_keybord"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/cluster/rpc/grpc_client"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/log"
	"github.com/tristan-club/wizard/pkg/tstore"
	"github.com/tristan-club/wizard/pkg/util"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PreCheckResult struct {
	shouldHandle bool
	cmdId        string
	isCmd        bool
	us           *userstate.UserState
	handle       flow.TGFlowHandler
}

func (p *PreCheckResult) ShouldHandle() bool {
	return p.shouldHandle
}

type TGMgr struct {
	controllerMgr controller_pb.ControllerServiceClient
	botApi        *tgbotapi.BotAPI
	botName       string
	cmdList       []string
	cmdHandler    map[string]flow.TGFlowHandler
	cmdDesc       map[string]string
	cmdParser     []func(u *tgbotapi.Update) string
}

func NewTGMgr(controllerSvc, tStoreSvc string, presetCmdIdList []string) (*TGMgr, error) {
	controllerConn, err := grpc_client.Start(controllerSvc)
	if err != nil {
		return nil, fmt.Errorf("init controller conn error %s", err.Error())
	}

	if err := tstore.InitTStore(tStoreSvc); err != nil {
		return nil, fmt.Errorf("init tstore mgr error %s", err.Error())
	}

	tgMgr := &TGMgr{
		controllerMgr: controller_pb.NewControllerServiceClient(controllerConn),
		cmdList:       presetCmdIdList,
		cmdHandler:    map[string]flow.TGFlowHandler{},
		cmdDesc:       map[string]string{},
		cmdParser:     make([]func(u *tgbotapi.Update) string, 0),
	}

	for _, cmdId := range presetCmdIdList {
		tgMgr.cmdDesc[cmdId] = cmd.GetCmdDesc(cmdId)
		handler := cmdhandler.GetCmdHandler(cmdId)
		if handler == nil {
			if config.EnvIsDev() {
				continue
			}
			return nil, fmt.Errorf("invalid preset command config %s, handler is nil", cmdId)
		}
		tgMgr.cmdHandler[cmdId] = handler
		if handler.GetCmdParser() != nil {
			tgMgr.cmdParser = append(tgMgr.cmdParser, handler.GetCmdParser())
		}

		if cmdId == cmd.CmdCreateEnvelope {
			tgMgr.cmdParser = append(tgMgr.cmdParser, cmdhandler.GetCmdHandler(cmd.CmdOpenEnvelope).GetCmdParser())
			tgMgr.cmdHandler[cmd.CmdOpenEnvelope] = cmdhandler.GetCmdHandler(cmd.CmdOpenEnvelope)
		}
	}
	return tgMgr, nil
}

func (t *TGMgr) RegisterCmd(cmdId, desc string, handler flow.TGFlowHandler) error {
	if cmdId == "" || desc == "" || handler == nil {
		return fmt.Errorf("invalid cmd config, cmdId %s", cmdId)
	}

	t.cmdList = append(t.cmdList, cmdId)
	t.cmdDesc[cmdId] = desc
	t.cmdHandler[cmdId] = handler
	if handler.GetCmdParser() != nil {
		t.cmdParser = append(t.cmdParser, handler.GetCmdParser())
	}

	return nil
}

func (t *TGMgr) ListenTGUpdate(botToken, webhookUrl, httpAddr string) error {
	var err error
	t.botApi, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("init bot error %s", err.Error())
	}

	botInfo, err := t.botApi.GetMe()
	if err != nil {
		return fmt.Errorf("get bot info error %s", err.Error())
	}

	t.botName = botInfo.UserName

	var updates tgbotapi.UpdatesChannel

	if webhookUrl != "" {

		if httpAddr == "" {
			return fmt.Errorf("invalid webhook http listen addr")
		}

		whUrl, err := url.Parse(webhookUrl)
		if err != nil {
			return fmt.Errorf("parse webhook url error %s", err.Error())
		}
		var tgListen string
		if whUrl.RequestURI() == "/" {
			tgListen = whUrl.RequestURI() + t.botName
		} else {
			tgListen = whUrl.RequestURI() + "/" + t.botName
		}

		updates = t.botApi.ListenForWebhook(tgListen)

		whConfig, err := tgbotapi.NewWebhook(fmt.Sprintf("%s/%s", webhookUrl, t.botName))
		if err != nil {
			return fmt.Errorf("new webhook config error %s", err.Error())
		}
		_, err = t.botApi.Request(whConfig)
		if err != nil {
			return fmt.Errorf("request webhook error %s", err.Error())
		}

		whInfo, err := t.botApi.GetWebhookInfo()
		if err != nil {
			return fmt.Errorf("get webhook info error %s", err.Error())
		}

		if whInfo.LastErrorDate != 0 {
			lastErrorDate := time.Unix(int64(whInfo.LastErrorDate), 0)
			log.Error().Fields(map[string]interface{}{
				"action":          "start telegram bot latest error",
				"bot name":        t.botName,
				"last error date": lastErrorDate,
				"error":           fmt.Errorf("telegram callback failed: %s", whInfo.LastErrorMessage),
			}).Send()
		}

		go func() {
			if err := http.ListenAndServe(httpAddr, nil); err != nil {
				log.Error().Fields(map[string]interface{}{"action": "init tg webhook http svc error", "error": err.Error()}).Send()
				panic(err)
			}
		}()

		log.Info().Fields(map[string]interface{}{"action": "success init tg bot webhook", "botName": t.botName, "webHookUrl": webhookUrl, "httpUrl": httpAddr, "listenPath": tgListen}).Send()

	} else {

		dw := tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true}
		_, err := t.botApi.Request(dw)
		if err != nil {
			return fmt.Errorf("delete webhook config error %s", err.Error())
		}

		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		//b.Debug = true
		updates = t.botApi.GetUpdatesChan(u)

		log.Info().Fields(map[string]interface{}{"action": "success init tg bot websocket", "botName": t.botName}).Send()
	}

	if len(t.cmdHandler) > 0 {
		var botCmdList []tgbotapi.BotCommand
		for _, v := range t.cmdList {
			botCmdList = append(botCmdList, tgbotapi.BotCommand{
				Command:     v,
				Description: t.cmdDesc[v],
			})
		}
		if _, err = t.botApi.Request(tgbotapi.SetMyCommandsConfig{Commands: botCmdList}); err != nil {
			return fmt.Errorf("bot register cmd error %s", err)
		}
	}

	for update := range updates {
		go t.handleTGUpdate(&update, nil)
	}

	return nil
}

func (t *TGMgr) InjectBotApi(botApi *tgbotapi.BotAPI, botName string) error {
	if botApi == nil || botName == "" {
		return fmt.Errorf("invalid bot api inject")
	}
	t.botApi = botApi
	t.botName = botName
	return nil
}

func (t *TGMgr) HandleTGUpdate(update *tgbotapi.Update, result *PreCheckResult) {
	go t.handleTGUpdate(update, result)
}

func (t *TGMgr) handleTGUpdate(update *tgbotapi.Update, preCheckResult *PreCheckResult) (shouldHandle bool) {
	err := t.handle(update, preCheckResult)
	if err != nil {

		openId := strconv.FormatInt(update.SentFrom().ID, 10)
		var isBusinessError bool

		if resetStateError := userstate.ResetState(openId); resetStateError != nil {
			log.Error().Fields(map[string]interface{}{"action": "reset user state error", "error": resetStateError.Error()}).Send()
		}

		var content string
		us, _ := userstate.GetState(openId, nil)
		herr, ok := err.(he.Error)
		if ok {
			if herr.ErrorType() == he.BusinessError {
				isBusinessError = true
				content = fmt.Sprintf(herr.Msg())
			} else {
				content = fmt.Sprintf("code:%d; error:%s; detail:%s", herr.Code(), herr.Msg(), herr.Error())
			}

		} else {
			content = err.Error()
		}

		if isBusinessError {
			log.Info().Fields(map[string]interface{}{"action": "got business error", "detail": content, "us": us}).Send()
		} else {
			log.Error().Fields(map[string]interface{}{"action": "got Server error", "detail": content, "us": us}).Send()
		}

		message := tgbotapi.NewMessage(update.FromChat().ID, content)
		if update.Message != nil {
			message.ReplyToMessageID = update.Message.MessageID
		}

		_, err = t.botApi.Send(message)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "send error message to user", "error": err.Error(), "content": content}).Send()
			message.Text = text.ServerError
			_, err = t.botApi.Send(message)
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "send server error", "error": err.Error()}).Send()
			}
		}
	}
	return shouldHandle
}

func (t *TGMgr) CheckShouldHandle(update *tgbotapi.Update) (pcr *PreCheckResult, err error) {

	var cmdId string
	var isCmd bool
	pcr = &PreCheckResult{}

	var message *tgbotapi.Message

	if update.CallbackQuery != nil {
		message = update.CallbackQuery.Message
	} else if update.Message != nil {
		if !update.FromChat().IsPrivate() && !update.Message.IsCommand() {
			return pcr, nil
		}
		message = update.Message
	} else {
		log.Info().Msgf("unknown message %s", util.FastMarshal(update))
		return pcr, nil
	}

	userId := strconv.FormatInt(update.SentFrom().ID, 10)

	us, herr := userstate.GetState(userId, nil)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "get user state error", "error": herr.Error()}).Send()
		return pcr, herr
	}

	if update.Message != nil {
		if update.Message.NewChatMembers != nil || update.Message.LeftChatMember != nil {
			return pcr, nil
		}
		if message.IsCommand() {
			msgText := message.Text
			if len(msgText) > 0 && msgText[0:1] == "/" {
				// this is a group msg, check if msg is sent to this botguai'wo
				if !message.Chat.IsPrivate() {
					if strings.Contains(msgText, t.botName) {
						msgText = strings.Replace(msgText, fmt.Sprintf("@%s", t.botName), " ", 1)
					} else {
						_cmdId, ok := parseCmd(msgText)
						if ok {
							if _, ok := t.cmdDesc[_cmdId]; ok {
								cmdId = _cmdId
							} else {
								log.Info().Fields(map[string]interface{}{"action": "got unknown cmd", "cmd": msgText}).Send()
								return pcr, nil
							}
						}
					}
				}
				_cmdId, ok := parseCmd(msgText)
				if ok {
					cmdId = _cmdId
				}
			}
			if cmdId == "" {
				log.Error().Fields(map[string]interface{}{
					"action":  "parse cmd",
					"payload": util.FastMarshal(update),
				}).Send()
				return pcr, he.NewBusinessError(he.CodeInvalidCmd, "", nil)
			}
			isCmd = true
			//if err := userstate.ResetState(userId); err != nil {
			//	log.Error().Fields(map[string]interface{}{"action": "reset user state", "error": err.Error()}).Send()
			//	return pcr, err
			//}

		} else {
			if us.CurrentCommand != "" && us.CurrentCommand != userstate.CmdNone {
				cmdId = us.CurrentCommand
			}
		}
	} else if update.CallbackQuery != nil {
		for _, fn := range t.cmdParser {
			if parsedCmdId := fn(update); parsedCmdId != "" {
				cmdId = parsedCmdId
				isCmd = true
				break
			}
		}

		if cmdId == "" {
			cmdId = us.CurrentCommand
		}

	}

	if cmdId == "" || cmdId == userstate.CmdNone {
		return pcr, nil
	}

	cmdHandler := t.cmdHandler[cmdId]
	if cmdHandler == nil {
		return pcr, nil
	}

	return &PreCheckResult{
		shouldHandle: true,
		cmdId:        cmdId,
		isCmd:        isCmd,
		us:           us,
		handle:       cmdHandler,
	}, nil
}

func (t *TGMgr) handle(update *tgbotapi.Update, preCheckResult *PreCheckResult) (err error) {

	if preCheckResult == nil {
		preCheckResult, err = t.CheckShouldHandle(update)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "parse update error", "error": err.Error()}).Send()
			return err
		}
		if !preCheckResult.ShouldHandle() {
			log.Info().Fields(map[string]interface{}{"action": "get unknown update, return", "updateId": update.UpdateID}).Send()
			return nil
		}
	}

	//var message *tgbotapi.Message
	//
	//if update.CallbackQuery != nil {
	//	message = update.CallbackQuery.Message
	//} else if update.Message != nil {
	//	message = update.Message
	//} else {
	//	log.Error().Msgf("unknown message %s", util.FastMarshal(update))
	//	return nil
	//}

	userId := strconv.FormatInt(update.SentFrom().ID, 10)

	cmdId := preCheckResult.cmdId
	us := preCheckResult.us
	isCmd := preCheckResult.isCmd
	cmdHandler := preCheckResult.handle

	c, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()
	ctx := &tcontext.Context{
		CmdId:   cmdId,
		Context: c,
		U:       update,
		CM:      t.controllerMgr,
		BotApi:  t.botApi,
		BotName: t.botName,
	}

	requester := &controller_pb.Requester{
		RequesterOpenId:       userId,
		RequesterOpenType:     pconst.PlatformTg,
		RequesterOpenNickname: ctx.GetNickname(),
		RequesterOpenUserName: ctx.GetUserName(),
	}

	if !update.FromChat().IsPrivate() {
		requester.RequesterChannelId = strconv.FormatInt(update.FromChat().ID, 10)
	}

	if isCmd {

		if err = userstate.ResetState(userId); err != nil {
			log.Error().Fields(map[string]interface{}{"action": "reset user state", "error": err.Error()}).Send()
			return err
		}

		b, err := proto.Marshal(requester)
		if err != nil {
			return he.NewServerError(he.CodeMarshalError, "", err)
		}

		requestStr := base64.StdEncoding.EncodeToString(b)
		md := metadata.Pairs("requester", requestStr)
		ctx.Context = metadata.NewOutgoingContext(c, md)
		ctx.Requester = requester

		getUserResp, err := t.controllerMgr.GetUser(ctx.Context, &controller_pb.GetUserReq{
			OpenId:   requester.RequesterOpenId,
			OpenType: requester.RequesterOpenType,
		})
		if err != nil {
			return he.NewServerError(he.CodeWalletRequestError, "", err)
		} else if getUserResp.CommonResponse.Code != he.Success {
			if getUserResp.CommonResponse.Code == pconst.CODE_USER_NOT_EXIST {
				var heErr he.Error

				if cmdId != cmd.CmdStart {
					if ctx.U.FromChat().IsPrivate() {
						_, heErr = ctx.Reply(ctx.U.SentFrom().ID, fmt.Sprintf(text.UserNoInit, ctx.GetNickNameMDV2()), nil, true)

					} else {
						var replyMsg tgbotapi.Message
						inlineKeyboard, deadlineTime := inline_keybord.NewForwardCreateKeyBoard(ctx)
						if replyMsg, heErr = ctx.Reply(ctx.U.SentFrom().ID, fmt.Sprintf(text.UserNoInit, ctx.GetNickNameMDV2()), inlineKeyboard, true); heErr == nil {
							inline_keybord.DeleteDeadKeyboard(ctx, deadlineTime, &replyMsg)
						}
					}
					return heErr
					//return he.NewBusinessError(he.CodeUserNotInit, text.UserNoInit)
				}

			} else {
				return he.NewServerError(int(getUserResp.CommonResponse.Code), "", fmt.Errorf(getUserResp.CommonResponse.Message))
			}
		} else {
			requester.RequesterUserNo = getUserResp.Data.UserNo
			requester.RequesterDefaultAddress = getUserResp.Data.DefaultAccountAddr

			if getUserResp.Data.OpenNickname != ctx.GetNickname() || getUserResp.Data.OpenUsername != ctx.GetUserName() {
				updateUserReq := &controller_pb.UpdateUserReq{
					UserNo:     getUserResp.Data.UserNo,
					OpenId:     "",
					OpenType:   0,
					IsOpenInit: false,
					Username:   ctx.GetUserName(),
					Nickname:   ctx.GetNickname(),
				}
				updateUserResp, err := t.controllerMgr.UpdateUser(ctx.Context, updateUserReq)
				if err != nil {
					log.Error().Fields(map[string]interface{}{"action": "call controller error", "error": err.Error(), "req": updateUserReq}).Send()
				} else if updateUserResp.CommonResponse.Code != he.Success {
					log.Error().Fields(map[string]interface{}{"action": "update user error", "error": updateUserResp, "req": updateUserReq}).Send()
				}
			}

		}

		if herr := userstate.InitState(userId, cmdId, getUserResp.Data.UserNo, getUserResp.Data.DefaultAccountAddr); herr != nil {
			return herr
		}
		ctx.CurrentState = userstate.StateNone
	} else {
		requester.RequesterUserNo = us.UserNo
		requester.RequesterDefaultAddress = us.DefaultAddress
		if requester.RequesterUserNo == "" || requester.RequesterDefaultAddress == "" {
			return he.NewBusinessError(he.CodeUserNotInit, "", nil)
		}
		ctx.CurrentState = us.CurrentState
	}

	ctx.Requester = requester
	b, err := proto.Marshal(requester)
	if err != nil {
		return he.NewServerError(he.CodeMarshalError, "", err)
	}

	requestStr := base64.StdEncoding.EncodeToString(b)
	md := metadata.Pairs("requester", requestStr)
	ctx.Context = metadata.NewOutgoingContext(c, md)

	return cmdHandler.Handle(ctx)

}

func parseCmd(content string) (string, bool) {
	reg, err := regexp.Compile("[^ /]+")

	if err != nil {
		log.Error().Msgf("compile error: %s", err)
		return "", false
	}

	argv := reg.FindAllString(content, -1)
	var cmdId string
	//tcmd := bot_pb.Command{
	//	Cmd:  "",
	//	Args: nil,
	//}

	if len(argv) > 0 {
		cmdId = argv[0]
	}

	return cmdId, true
}
