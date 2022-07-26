package dcmgr

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/bot-wizard/cmd"
	"github.com/tristan-club/bot-wizard/config"
	"github.com/tristan-club/bot-wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/handler"
	"github.com/tristan-club/bot-wizard/handler/discordhandler/handler/cmdhandler"
	"github.com/tristan-club/bot-wizard/handler/text"
	"github.com/tristan-club/bot-wizard/handler/userstate"
	"github.com/tristan-club/bot-wizard/pconst"
	"github.com/tristan-club/bot-wizard/pkg/cluster/rpc/grpc_client"
	"github.com/tristan-club/bot-wizard/pkg/dingding"
	he "github.com/tristan-club/bot-wizard/pkg/error"
	"github.com/tristan-club/bot-wizard/pkg/log"
	"github.com/tristan-club/bot-wizard/pkg/tstore"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"os"
	"time"
)

type PreCheckResult struct {
	shouldHandle bool
	cmdId        string
	isCmd        bool
	us           *userstate.UserState
	handler      *handler.DiscordCmdHandler
}

func (p *PreCheckResult) ShouldHandle() bool {
	return p.shouldHandle
}

type Manager struct {
	controllerMgr controller_pb.ControllerServiceClient
	s             *discordgo.Session
	botName       string
	cmdList       []string
	cmdHandler    map[string]*handler.DiscordCmdHandler
	cmdDesc       map[string]string
	//cmdParser     []func(u *tgbotapi.Update) string
}

func NewMgr(controllerSvc, tStoreSvc string, presetCmdIdList []string) (*Manager, error) {
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
		cmdDesc:       map[string]string{},
	}

	for _, cmdId := range presetCmdIdList {
		mgr.cmdDesc[cmdId] = cmd.GetCmdDesc(cmdId)
		h := cmdhandler.GetCmdHandler(cmdId)
		if h == nil {
			if config.EnvIsDev() {
				continue
			}
			return nil, fmt.Errorf("invalid preset command config %s, handler is nil", cmdId)
		}
		mgr.cmdHandler[cmdId] = h
	}
	return mgr, nil
}
func (t *Manager) SetCmd(handler *handler.DiscordCmdHandler) {
	if handler == nil {
		log.Error().Fields(map[string]interface{}{"action": "empty handler set"}).Send()
		return
	}

	t.cmdList = append(t.cmdList, handler.ApplicationCommand.Name)
	t.cmdDesc[handler.ApplicationCommand.Name] = handler.ApplicationCommand.Description
	t.cmdHandler[handler.ApplicationCommand.Name] = handler

}

func (t *Manager) RegisterCmd() error {
	if len(t.cmdHandler) > 0 && os.Getenv("RESET_DISCORD_CMD") != "1" {
		for _, cmdHandler := range t.cmdHandler {
			_, err := t.s.ApplicationCommandCreate(t.s.State.User.ID, "", cmdHandler.ApplicationCommand)
			log.Info().Fields(map[string]interface{}{"action": "register cmd", "cmdId": cmdHandler.ApplicationCommand.Name}).Send()
			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "register cmd error", "error": err.Error()}).Send()
			}
		}
	}
	return nil
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
	go t.handleTGUpdate(i, result)
}

func (t *Manager) handleTGUpdate(i *discordgo.InteractionCreate, pcr *PreCheckResult) (shouldHandle bool) {
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

		err = t.s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   uint64(discordgo.MessageFlagsEphemeral),
			}})
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "send error message to user", "error": err.Error(), "content": content}).Send()
			err = t.s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: text.ServerError,
					Flags:   uint64(discordgo.MessageFlagsEphemeral),
				}})

			if err != nil {
				log.Error().Fields(map[string]interface{}{"action": "send server error", "error": err.Error()}).Send()
			}
		}

	}
	return shouldHandle
}

func (t *Manager) CheckShouldHandle(i *discordgo.InteractionCreate) (pcr *PreCheckResult, err error) {

	pcr = &PreCheckResult{}

	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

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
	}

	requester := &controller_pb.Requester{
		RequesterUserNo:         "",
		RequesterOpenId:         ctx.GetFromId(),
		RequesterOpenType:       pconst.PlatformDiscord,
		RequesterOpenNickname:   ctx.GetNickname(),
		RequesterOpenUserName:   ctx.GetUserName(),
		RequesterChannelId:      ctx.GetGroupChannelId(),
		RequesterDefaultAddress: "",
	}

	c, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	b, err := proto.Marshal(requester)
	if err != nil {
		return he.NewServerError(he.CodeMarshalError, "", err)
	}

	requestStr := base64.StdEncoding.EncodeToString(b)
	md := metadata.Pairs("requester", requestStr)
	ctx.Context = metadata.NewOutgoingContext(c, md)
	ctx.Requester = requester

	cmdId := ctx.IC.Interaction.ApplicationCommandData().Name

	getUserResp, err := ctx.CM.GetUser(ctx.Context, &controller_pb.GetUserReq{
		OpenId:   requester.RequesterOpenId,
		OpenType: requester.RequesterOpenType,
	})
	if err != nil {
		return he.NewServerError(he.CodeWalletRequestError, "", err)
	} else if getUserResp.CommonResponse.Code != he.Success {
		if getUserResp.CommonResponse.Code == pconst.CODE_USER_NOT_EXIST {

			if cmdId != cmd.CmdStart {
				err = ctx.ReplyDmWithGroupForward("", "", "You do not have an account yet, use start command to create account")
				if err != nil {
					return he.NewServerError(he.CodeBotSendMsgError, "", err)
				}
			}

		} else {
			return he.NewServerError(int(getUserResp.CommonResponse.Code), "", fmt.Errorf(getUserResp.CommonResponse.Message))
		}
	} else {
		requester.RequesterUserNo = getUserResp.Data.UserNo
		requester.RequesterDefaultAddress = getUserResp.Data.DefaultAccountAddr
	}

	return pcr.handler.Handler(ctx)
}
