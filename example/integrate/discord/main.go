package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/handler/discordhandler/handler/cmdhandler"
	"github.com/tristan-club/wizard/manager/dcmgr"
	"github.com/tristan-club/wizard/pkg/dingding"
	"github.com/tristan-club/wizard/pkg/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	time.Local = time.FixedZone("UTC", 0)

	b, err := discordgo.New("Bot " + os.Getenv("DC_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}
	b.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = b.Open()

	if err != nil {
		retryMax := 3
		for i := 1; i <= retryMax; i++ {
			time.Sleep(time.Duration(i*15) * time.Second)
			if err = b.Open(); err == nil {
				break
			} else {
				if i == retryMax {
					content := fmt.Sprintf(`### Discord 机器人告警,链接重试 %d 次失败，请查看`, i)
					robot := dingding.NewRobot(config.GetDingDingToken(), "", "", "")
					if err := robot.SendMarkdownMessage("## Discord Wizard", content, nil, true); err != nil {
						log.Error().Msgf("send dingding msg error:%s", err)
					}
					log.Error().Fields(map[string]interface{}{
						"action": "start discord bot",
						"error":  err,
					}).Send()
					panic(err)
				}
			}
		}
	}

	log.Info().Fields(map[string]interface{}{"action": "discord bot open session success", "botName": b.State.User.Username}).Send()

	dcMgr, err := dcmgr.NewMgr(os.Getenv("CONTROLLER_SERVICE"), os.Getenv("TSTORE_SERVICE"), cmd.GetCmdList())
	if err != nil {
		panic(err)
	}

	if err = dcMgr.InjectBotApi(b, b.State.User.Username); err != nil {
		panic(err)
	}

	b.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		pcr, err := dcMgr.CheckShouldHandle(i)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "parse ic error", "error": err.Error()}).Send()
		} else if !pcr.ShouldHandle() {
			log.Error().Fields(map[string]interface{}{"action": "get unknown message", "i": i.ApplicationCommandData()}).Send()
		} else {
			dcMgr.Handle(i, pcr)
		}

	})
	for _, v := range cmd.GetCmdList() {
		h := cmdhandler.GetCmdHandler(v)
		if h != nil {
			dcMgr.SetCmd(h)
		}
	}
	if err = dcMgr.RegisterCmd(); err != nil {
		panic(fmt.Sprintf("register cmd error %s", err.Error()))
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	b.Close()

}
