package dcmgr

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/kit/tstore"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/discordhandler/handler"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_change_pincode"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_envelope"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmd_submit_metamask"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmdhandler"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/cluster/rpc/grpc_client"
	"github.com/tristan-club/wizard/pkg/dingding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"os"
	"time"
)

type PreCheckResult struct {
	shouldHandle bool
	cmdId        string
	cid          *customid.CustomId
	isCmd        bool
	us           *userstate.UserState
	handler      *handler.DiscordCmdHandler
	payload      interface{}
}

func (p *PreCheckResult) ShouldHandle() bool {
	return p.shouldHandle
}

func (p *PreCheckResult) InjectPayload(payload interface{}) {
	p.payload = payload
}

type Manager struct {
	controllerMgr controller_pb.ControllerServiceClient
	s             *discordgo.Session
	botName       string
	cmdList       []string
	cmdHandler    map[string]*handler.DiscordCmdHandler
	customHandler map[int32]*handler.DiscordCmdHandler
	cmdDesc       map[string]string
	appId         string
	//cmdParser     []func(u *tgbotapi.Update) string
}

func NewMgr(controllerSvc, tStoreSvc string, presetCmdIdList []string, appId string) (*Manager, error) {
	controllerConn, err := grpc_client.Start(controllerSvc)
	if err != nil {
		return nil, fmt.Errorf("init controller conn error %s", err.Error())
	}

	if err := tstore.InitTStore(tStoreSvc); err != nil {
		return nil, fmt.Errorf("init tstore mgr error %s", err.Error())
	}

	mgr := &Manager{
		controllerMgr: controller_pb.NewControllerServiceClient(controllerConn),
		cmdList:       presetCmdIdList,
		cmdHandler:    map[string]*handler.DiscordCmdHandler{},
		customHandler: map[int32]*handler.DiscordCmdHandler{},
		cmdDesc:       map[string]string{},
		appId:         appId,
	}

	for _, cmdId := range presetCmdIdList {
		mgr.cmdDesc[cmdId] = cmd.GetCmdDesc(cmdId)
		h := cmdhandler.GetCmdHandler(cmdId)
		if h == nil {
			if config.EnvIsDev() {
				continue
			}
			return nil, fmt.Errorf("invalid preset command config %s, chandler is nil", cmdId)
		}
		mgr.cmdHandler[cmdId] = h
	}
	return mgr, nil
}

func (t *Manager) SetCmd(handler *handler.DiscordCmdHandler) {
	if handler == nil {
		log.Error().Fields(map[string]interface{}{"action": "empty chandler set"}).Send()
		return
	}

	t.cmdList = append(t.cmdList, handler.ApplicationCommand.Name)
	t.cmdDesc[handler.ApplicationCommand.Name] = handler.ApplicationCommand.Description
	t.cmdHandler[handler.ApplicationCommand.Name] = handler

}

func (t *Manager) DeleteCmd(cmdId string) {
	if cmdId == "" {
		return
	}

	var key = -1
	for i, cmdStr := range t.cmdList {
		if cmdStr == cmdId {
			key = i
			break
		}
	}
	if key < 0 {
		// not existed cmdId
		return
	}

	t.cmdList = append(t.cmdList[:key], t.cmdList[key:]...)
	delete(t.cmdDesc, cmdId)
	delete(t.cmdHandler, cmdId)
}

func (t *Manager) GetCmdList() []*discordgo.ApplicationCommand {
	if t.cmdHandler == nil {
		return nil
	}
	var resp []*discordgo.ApplicationCommand
	for _, v := range t.cmdList {
		resp = append(resp, t.cmdHandler[v].ApplicationCommand)
	}
	return resp
}

func (t *Manager) RegisterCmd() error {
	if len(t.cmdHandler) > 0 && os.Getenv("RESET_DISCORD_CMD") == "1" {
		var handleList []*discordgo.ApplicationCommand

		for _, cmdHandler := range t.cmdHandler {
			handleList = append(handleList, cmdHandler.ApplicationCommand)
		}
		_, err := t.s.ApplicationCommandBulkOverwrite(t.s.State.User.ID, "", handleList)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Manager) RegisterCustomHandler(cid *customid.CustomId, h *handler.DiscordCmdHandler) {
	t.customHandler[cid.GetCustomType()] = h
}

func (t *Manager) SetCustomHandler(customType int32, h *handler.DiscordCmdHandler) {
	t.customHandler[customType] = h
}

func (t *Manager) InjectBotApi(botApi *discordgo.Session, botName string) error {
	if botApi == nil || botName == "" {
		return fmt.Errorf("invalid bot api inject")
	}
	t.s = botApi
	t.botName = botName
	return nil
}

func (t *Manager) ListenDCInteraction(botToken string) error {
	var err error
	t.s, err = discordgo.New("Bot " + botToken)
	if err != nil {
		return err
	}
	t.s.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	t.s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

	})

	t.s.AddHandler(func(s *discordgo.Session, i *discordgo.MessageCreate) {

	})
	if err != nil {
		retryMax := 3
		for i := 1; i <= retryMax; i++ {
			time.Sleep(time.Duration(i*15) * time.Second)
			if err = t.s.Open(); err == nil {
				break
			} else {
				if i == retryMax {
					token := config.GetDingDingToken()
					if token != "" {
						content := fmt.Sprintf(`### Discord 机器人告警,链接重试 %d 次失败，请查看`, i)
						robot := dingding.NewRobot(config.GetDingDingToken(), "", "", "")
						if err := robot.SendMarkdownMessage("## Wizard ", content, nil, true); err != nil {
							log.Error().Msgf("send dingding msg error:%s", err)
						}
					}

					log.Error().Fields(map[string]interface{}{
						"action": "start discord bot",
						"error":  err,
					}).Send()
					return err
				}
			}
		}
	}
	log.Info().Fields(map[string]interface{}{"action": "open discord connection success"}).Send()

	return t.RegisterCmd()

}

func (t *Manager) Handle(i *discordgo.InteractionCreate, result *PreCheckResult) {
	go t.handleUpdate(i, result)
}

func (t *Manager) CheckShouldHandle(i *discordgo.InteractionCreate) (pcr *PreCheckResult, err error) {

	pcr = &PreCheckResult{}

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		cmdHandler := t.cmdHandler[i.ApplicationCommandData().Name]
		if cmdHandler == nil {
			return pcr, nil
		}
		return &PreCheckResult{
			shouldHandle: true,
			cmdId:        i.ApplicationCommandData().Name,
			isCmd:        true,
			handler:      cmdHandler,
		}, nil

	case discordgo.InteractionMessageComponent:
		cid, ok := customid.ParseCustomId(i.MessageComponentData().CustomID)
		if !ok {
			return
		}
		pcr.cid = cid
		pcr.isCmd = true
		pcr.shouldHandle = true
		switch cid.GetCustomType() {
		case pconst.CustomIdOpenEnvelope:
			pcr.handler = cmd_envelope.OpenEnvelopeHandler
		case pconst.CustomIdChangePinCode:
			pcr.handler = cmd_change_pincode.Handler
		case pconst.CustomIdSubmitMetaMask:
			pcr.handler = cmd_submit_metamask.Handler
		}

		if pcr.handler == nil {
			pcr.handler, ok = t.customHandler[cid.GetCustomType()]
			if ok {
				pcr.isCmd = true
			} else {
				pcr.isCmd = false
				pcr.shouldHandle = false
			}
		}

		return
	default:
		return
	}

}

func (t *Manager) handleUpdate(i *discordgo.InteractionCreate, pcr *PreCheckResult) {

	ctx := dcontext.DefaultContext(t.s, i)

	if err := ctx.AckMsg(false); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "ack msg error", "error": err.Error(), "i": i}).Send()
		return
	}

	err := t.handle(i, pcr)
	if err != nil {
		var isBusinessError bool
		var content string
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
			log.Info().Fields(map[string]interface{}{"action": "got business error", "detail": content}).Send()
		} else {
			log.Error().Fields(map[string]interface{}{"action": "got Server error", "detail": content}).Send()
		}

		_, err = ctx.FollowUpReply(content)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "send error message to user", "error": err.Error(), "content": content}).Send()
			_, err = ctx.FollowUpReply(text.ServerError)
		}

	}
	return
}

func (t *Manager) handle(i *discordgo.InteractionCreate, pcr *PreCheckResult) (err error) {

	ctx := &dcontext.Context{
		IC:        i,
		Session:   t.s,
		CmdId:     pcr.cmdId,
		Context:   context.Background(),
		Requester: nil,
		BotName:   t.botName,
		CM:        t.controllerMgr,
		Cid:       pcr.cid,
		Payload:   pcr.payload,
	}

	requester := &controller_pb.Requester{
		RequesterUserNo:         "",
		RequesterOpenId:         ctx.GetFromId(),
		RequesterOpenType:       pconst.PlatformDiscord,
		RequesterOpenNickname:   ctx.GetNickname(),
		RequesterOpenUserName:   ctx.GetUserName(),
		RequesterChannelId:      ctx.GetGroupChannelId(),
		RequesterDefaultAddress: "",
		RequesterAppId:          t.appId,
	}

	c, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	b, err := proto.Marshal(requester)
	if err != nil {
		return he.NewServerError(pconst.CodeMarshalError, "", err)
	}

	requestStr := base64.StdEncoding.EncodeToString(b)
	md := metadata.Pairs("requester", requestStr)
	ctx.Context = metadata.NewOutgoingContext(c, md)
	ctx.Requester = requester

	cmdId := pcr.cmdId

	getUserResp, err := ctx.CM.GetUser(ctx.Context, &controller_pb.GetUserReq{
		OpenId:   requester.RequesterOpenId,
		OpenType: requester.RequesterOpenType,
	})
	if err != nil {
		return he.NewServerError(pconst.CodeWalletRequestError, "", err)
	} else if getUserResp.CommonResponse.Code != he.Success {
		if getUserResp.CommonResponse.Code == pconst.CODE_USER_NOT_EXIST {

			if cmdId != cmd.CmdStart && cmdId != cmd.CmdOpenEnvelope {
				// todo This is a temporary check, it  will be refactor further
				if cid := ctx.Cid; cid != nil && (cid.GetCustomType() == 91501 || cid.GetCustomType() == pconst.CustomIdOpenEnvelope) {

				} else {
					_, err = ctx.FollowUpReply("You do not have an account yet, use start command to create account")
					if err != nil {
						return he.NewServerError(pconst.CodeBotSendMsgError, "", err)
					}
					return nil
				}

			}

		} else {
			return he.NewServerError(int(getUserResp.CommonResponse.Code), "", fmt.Errorf(getUserResp.CommonResponse.Message))
		}
	} else {

		requester.RequesterUserNo = getUserResp.Data.UserNo
		requester.RequesterDefaultAddress = getUserResp.Data.DefaultAccountAddr

		avatarUrl := ctx.GetAvatarUrl()

		if getUserResp.Data.OpenUsername != ctx.GetUserName() || (getUserResp.Data.AppId == "" && requester.RequesterAppId != "") || getUserResp.Data.AvatarUrl != avatarUrl || getUserResp.Data.Code == "" {
			updateUserReq := &controller_pb.UpdateUserReq{
				UserNo:     getUserResp.Data.UserNo,
				OpenId:     ctx.GetFromId(),
				OpenType:   pconst.PlatformDiscord,
				IsOpenInit: false,
				Username:   ctx.GetUserName(),
				Nickname:   "",
				AppId:      requester.RequesterAppId,
				AvatarUrl:  avatarUrl,
				Code:       ctx.GetUser().Discriminator,
			}
			updateUserResp, err := t.controllerMgr.UpdateUser(ctx.Context, updateUserReq)
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "call controller error", "error": err.Error(), "req": updateUserReq}).Send()
			} else if updateUserResp.CommonResponse.Code != he.Success {
				log.Error().Fields(map[string]interface{}{"action": "update user error", "error": updateUserResp, "req": updateUserReq}).Send()
			}
		}
	}
	log.Info().Msgf("user %s begin %s cmd", ctx.Requester.RequesterOpenId, ctx.CmdId)

	defer func() {
		if panicErr := recover(); panicErr != nil {
			log.Error().Fields(map[string]interface{}{"action": "get panic error", "error": panicErr, "pcr": pcr, "ctx": ctx}).Send()
		}
	}()

	return pcr.handler.Handler(ctx)
}
