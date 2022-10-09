package tgmgr

import (
	"context"
	"encoding/base64"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow"
	"github.com/tristan-club/wizard/handler/tghandler/handler/commandhandler/cmdhandler"
	"github.com/tristan-club/wizard/handler/tghandler/inline_keybord"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/cluster/rpc/grpc_client"
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
	isCmd        bool
	cmdId        string
	cid          *customid.CustomId
	us           *userstate.UserState
	handler      flow.TGFlowHandler
	payload      interface{}
	cmdParam     []string
}

func (p *PreCheckResult) ShouldHandle() bool {
	return p.shouldHandle
}

func (p *PreCheckResult) InjectPayload(payload interface{}) {
	p.payload = payload
}

func (p *PreCheckResult) IsCmd() bool {
	return p.isCmd
}

func (p *PreCheckResult) CmdId() string {
	return p.cmdId
}

func (p *PreCheckResult) CmdParam() []string {
	return p.cmdParam
}

type TGMgr struct {
	controllerMgr controller_pb.ControllerServiceClient
	//widgetMgr     widget_pb.WidgetServiceClient
	botApi         *tgbotapi.BotAPI
	botName        string
	cmdList        []string
	cmdHandler     map[string]flow.TGFlowHandler
	customHandler  map[int32]flow.TGFlowHandler
	startHandler   []flow.TGFlowHandler
	cmdDesc        map[string]string
	cmdParser      []func(u *tgbotapi.Update) string
	appId          string
	availableChain []int32
}

func NewTGMgr(controllerSvc, tStoreSvc string, presetCmdIdList []string, appId string) (*TGMgr, error) {
	controllerConn, err := grpc_client.Start(controllerSvc)
	if err != nil {
		return nil, fmt.Errorf("init controller conn error %s", err.Error())
	}

	if err := tstore.InitTStore(tStoreSvc); err != nil {
		return nil, fmt.Errorf("init tstore mgr error %s", err.Error())
	}

	tgMgr := &TGMgr{
		controllerMgr:  controller_pb.NewControllerServiceClient(controllerConn),
		cmdList:        presetCmdIdList,
		cmdHandler:     map[string]flow.TGFlowHandler{},
		cmdDesc:        map[string]string{},
		customHandler:  map[int32]flow.TGFlowHandler{},
		cmdParser:      make([]func(u *tgbotapi.Update) string, 0),
		appId:          appId,
		availableChain: make([]int32, 0),
	}

	for _, cmdId := range presetCmdIdList {
		tgMgr.cmdDesc[cmdId] = cmd.GetCmdDesc(cmdId)
		handler := cmdhandler.GetCmdHandler(cmdId)
		if handler == nil {
			if config.EnvIsDev() {
				continue
			}
			return nil, fmt.Errorf("invalid preset command config %s, chandler is nil", cmdId)
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
func (t *TGMgr) AddAvailableChain(chainTypeList []int32) {
	t.availableChain = append(t.availableChain, chainTypeList...)
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

func (t *TGMgr) RegisterCustomHandler(cid *customid.CustomId, h flow.TGFlowHandler) {
	t.customHandler[cid.GetCustomType()] = h
}

func (t *TGMgr) RegisterStartHandler(h flow.TGFlowHandler) {
	t.startHandler = append(t.startHandler, h)
}

func (t *TGMgr) EnablePresetCmd(cmdIdList []string) {

	for _, cmdId := range cmdIdList {
		if handler := cmdhandler.GetCmdHandler(cmdId); handler != nil {
			t.cmdList = append(t.cmdList, cmdId)
			t.cmdDesc[cmdId] = cmd.GetCmdDesc(cmdId)
			t.cmdHandler[cmdId] = handler
			if handler.GetCmdParser() != nil {
				t.cmdParser = append(t.cmdParser, handler.GetCmdParser())
			}
		} else {
			log.Error().Fields(map[string]interface{}{"action": "invalid enbale preset cmd", "cmdId": cmdId}).Send()
		}
	}
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
		update := update

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

func (t *TGMgr) handleTGUpdate(update *tgbotapi.Update, preCheckResult *PreCheckResult) {
	err := t.handle(update, preCheckResult)
	if err != nil {
		var isMarkdown bool
		ctx := tcontext.DefaultContext(update, t.botApi)
		openId := strconv.FormatInt(update.SentFrom().ID, 10)
		var isBusinessError bool

		if resetStateError := userstate.ResetState(openId); resetStateError != nil {
			log.Error().Fields(map[string]interface{}{"action": "reset user state error", "error": resetStateError.Error()}).Send()
		}

		var content string
		var isGroupMsg bool
		us, _ := userstate.GetState(openId, nil)
		herr, ok := err.(he.Error)
		if ok {
			if strings.Contains(herr.Error(), "bot can't initiate conversation with a user") {
				content = fmt.Sprintf(text.ForbiddenError, ctx.GetNickNameMDV2())
				isMarkdown = true
				isGroupMsg = true
			} else if herr.ErrorType() == he.BusinessError {
				isBusinessError = true
				content = fmt.Sprintf(herr.Msg())
			} else {
				content = fmt.Sprintf("code: %d; error: %s; detail: %s", herr.Code(), herr.Msg(), herr.Error())
			}

		} else {
			content = err.Error()
		}

		if isBusinessError {
			log.Info().Fields(map[string]interface{}{"action": "got business error", "detail": content, "us": us}).Send()
		} else {
			log.Error().Fields(map[string]interface{}{"action": "got Server error", "detail": content, "us": us}).Send()
		}

		msg := &tgbotapi.Message{}

		if !isGroupMsg {
			msg, herr = ctx.Send(update.SentFrom().ID, content, nil, isMarkdown, false)
			if herr != nil {
				log.Warn().Fields(map[string]interface{}{"action": "bot send error msg error", "error": herr, "ctx": ctx, "content": content}).Send()
				msg, herr = ctx.Send(update.SentFrom().ID, text.ServerError, nil, isMarkdown, false)
				if herr != nil {
					log.Error().Fields(map[string]interface{}{"action": "send server error", "error": herr}).Send()
				}
			}
		} else {

			*msg, herr = ctx.Reply(update.FromChat().ID, content, nil, isMarkdown)
			if herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "send error message to user", "error": herr, "content": content}).Send()
				msg, herr = ctx.Send(update.FromChat().ID, text.ServerError, nil, isMarkdown, false)
				if herr != nil {
					log.Error().Fields(map[string]interface{}{"action": "send server error", "error": herr}).Send()
				}
			}
		}

		if msg != nil && msg.MessageID != 0 {
			ctx.SetDeadlineMsg(msg.Chat.ID, msg.MessageID, pconst.COMMON_MSG_DEADLINE)
		}
	}
	return
}

func (t *TGMgr) CheckShouldHandle(update *tgbotapi.Update) (pcr *PreCheckResult, err error) {

	var cmdId string
	var isCmd bool
	var cid *customid.CustomId
	pcr = &PreCheckResult{}
	var cmdParam []string

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
						_cmdId, args, ok := parseCmd(msgText)
						cmdParam = args
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
				_cmdId, args, ok := parseCmd(msgText)
				cmdParam = args
				if ok {
					cmdId = _cmdId
				}
			}
			if cmdId == "" {
				log.Error().Fields(map[string]interface{}{
					"action":  "parse cmd",
					"payload": util.FastMarshal(update),
				}).Send()
				return pcr, he.NewBusinessError(pconst.CodeInvalidCmd, "", nil)
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
			var ok bool
			cid, ok = customid.ParseCustomId(update.CallbackData())
			if ok {
				pcr.cid = cid
			}
		}

		if cmdId == "" && cid == nil {
			cmdId = us.CurrentCommand
		}

	}

	if (cmdId == "" || cmdId == userstate.CmdNone) && cid == nil {
		return pcr, nil
	}

	if cmdId != "" {

		var cmdHandler flow.TGFlowHandler
		if cmdId == cmd.CmdStart && len(t.startHandler) > 0 {
			for _, h := range t.startHandler {
				if h.GetCmdParser()(update) != "" {
					cmdHandler = h
					break
				}
			}
		}

		if cmdHandler == nil {
			cmdHandler = t.cmdHandler[cmdId]
		}

		if cmdHandler == nil {
			return pcr, nil
		} else {
			return &PreCheckResult{
				shouldHandle: true,
				cmdId:        cmdId,
				isCmd:        isCmd,
				us:           us,
				handler:      cmdHandler,
				cmdParam:     cmdParam,
			}, nil
		}
	} else {
		customHandler := t.customHandler[cid.GetCustomType()]
		if customHandler == nil {
			return pcr, nil
		} else {
			return &PreCheckResult{
				shouldHandle: true,
				cmdId:        "",
				isCmd:        false,
				us:           us,
				handler:      customHandler,
				cmdParam:     cmdParam,
			}, nil
		}
	}

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

	//log.Info().Fields(map[string]interface{}{"action": "get pcr", "prc": preCheckResult}).Send()

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
	cmdHandler := preCheckResult.handler
	cid := preCheckResult.cid

	c, cancel := context.WithTimeout(context.Background(), time.Second*600)
	defer cancel()
	ctx := &tcontext.Context{
		CmdId:    cmdId,
		Context:  c,
		U:        update,
		CM:       t.controllerMgr,
		BotApi:   t.botApi,
		BotName:  t.botName,
		Payload:  preCheckResult.payload,
		CmdParam: preCheckResult.cmdParam,
	}

	requester := &controller_pb.Requester{
		RequesterOpenId:       userId,
		RequesterOpenType:     pconst.PlatformTg,
		RequesterOpenNickname: ctx.GetNickname(),
		RequesterOpenUserName: ctx.GetUserName(),
		RequesterAppId:        t.appId,
	}

	if !update.FromChat().IsPrivate() {
		requester.RequesterChannelId = strconv.FormatInt(update.FromChat().ID, 10)
	}

	if isCmd || cid != nil {

		if err = userstate.ResetState(userId); err != nil {
			log.Error().Fields(map[string]interface{}{"action": "reset user state", "error": err.Error()}).Send()
			return err
		}

		b, err := proto.Marshal(requester)
		if err != nil {
			return he.NewServerError(pconst.CodeMarshalError, "", err)
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
			return he.NewServerError(pconst.CodeWalletRequestError, "", err)
		} else if getUserResp.CommonResponse.Code != he.Success {
			if getUserResp.CommonResponse.Code == pconst.CODE_USER_NOT_EXIST || (getUserResp.Data != nil && getUserResp.Data.DefaultAccountAddr == "") {
				var heErr he.Error

				if cmdId != cmd.CmdStart && !(cmdId == cmd.CmdStart && len(preCheckResult.cmdParam) == 1 && preCheckResult.cmdParam[0] == pconst.DefaultDeepLinkStart) {
					if ctx.U.FromChat().IsPrivate() {
						_, heErr = ctx.Reply(ctx.U.SentFrom().ID, fmt.Sprintf(text.UserNoInitInPrivate, ctx.GetNickNameMDV2()), nil, true)

					} else {
						var replyMsg tgbotapi.Message
						inlineKeyboard, deadlineTime := inline_keybord.NewForwardCreateKeyBoard(ctx)
						if replyMsg, heErr = ctx.Reply(ctx.U.SentFrom().ID, fmt.Sprintf(text.UserNoInit, ctx.GetNickNameMDV2()), inlineKeyboard, true); heErr == nil {
							inline_keybord.DeleteDeadKeyboard(ctx, deadlineTime, &replyMsg)
						}
					}
					return heErr
					//return he.NewBusinessError(pconst.CodeUserNotInit, text.UserNoInit)
				}

			} else {
				return he.NewServerError(int(getUserResp.CommonResponse.Code), "", fmt.Errorf(getUserResp.CommonResponse.Message))
			}
		} else {
			requester.RequesterUserNo = getUserResp.Data.UserNo
			requester.RequesterDefaultAddress = getUserResp.Data.DefaultAccountAddr
			requester.MetamaskAddress = getUserResp.Data.MetamaskAddress

			if getUserResp.Data.OpenNickname != ctx.GetNickname() || getUserResp.Data.OpenUsername != ctx.GetUserName() || (getUserResp.Data.AppId == "" && requester.RequesterAppId != "") {
				updateUserReq := &controller_pb.UpdateUserReq{
					UserNo:     getUserResp.Data.UserNo,
					OpenId:     "",
					OpenType:   0,
					IsOpenInit: false,
					Username:   ctx.GetUserName(),
					Nickname:   ctx.GetNickname(),
					AppId:      requester.RequesterAppId,
				}
				updateUserResp, err := t.controllerMgr.UpdateUser(ctx.Context, updateUserReq)
				if err != nil {
					log.Error().Fields(map[string]interface{}{"action": "call controller error", "error": err.Error(), "req": updateUserReq}).Send()
				} else if updateUserResp.CommonResponse.Code != he.Success {
					log.Error().Fields(map[string]interface{}{"action": "update user error", "error": updateUserResp, "req": updateUserReq}).Send()
				}
			}

		}

		if herr := userstate.InitState(userId, cmdId, requester.RequesterUserNo, requester.RequesterDefaultAddress); herr != nil {
			return herr
		}
		ctx.CurrentState = userstate.StateNone
	} else {
		requester.RequesterUserNo = us.UserNo
		requester.RequesterDefaultAddress = us.DefaultAddress
		if requester.RequesterUserNo == "" || requester.RequesterDefaultAddress == "" {
			return he.NewBusinessError(pconst.CodeUserNotInit, "", nil)
		}
		ctx.CurrentState = us.CurrentState
	}

	ctx.Requester = requester
	b, err := proto.Marshal(requester)
	if err != nil {
		return he.NewServerError(pconst.CodeMarshalError, "", err)
	}

	requestStr := base64.StdEncoding.EncodeToString(b)
	md := metadata.Pairs("requester", requestStr)
	ctx.Context = metadata.NewOutgoingContext(c, md)

	return cmdHandler.Handle(ctx)

}

func parseCmd(content string) (string, []string, bool) {
	reg, err := regexp.Compile("[^ /]+")

	if err != nil {
		log.Error().Msgf("compile error: %s", err)
		return "", nil, false
	}

	argv := reg.FindAllString(content, -1)
	var cmdId string
	var args []string
	//tcmd := bot_pb.Command{
	//	Cmd:  "",
	//	Args: nil,
	//}

	if len(argv) > 0 {
		cmdId = argv[0]
	}

	if len(argv) > 1 {
		args = argv[1:]
	}

	return cmdId, args, true
}
