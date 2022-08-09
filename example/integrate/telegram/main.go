package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/manager/tgmgr"
	"os"
	"time"
)

func main() {
	time.Local = time.FixedZone("UTC", 0)

	botApi, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	botInfo, err := botApi.GetMe()
	if err != nil {
		panic(err)
	}

	dw := tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true}
	_, err = botApi.Request(dw)
	if err != nil {
		panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	//b.Debug = true
	updates := botApi.GetUpdatesChan(u)

	tgBotMgr, err := tgmgr.NewTGMgr(os.Getenv("CONTROLLER_SERVICE"), os.Getenv("TSTORE_SERVICE"), cmd.GetCmdList())
	if err != nil {
		panic(err)
	}

	if err := tgBotMgr.InjectBotApi(botApi, botInfo.UserName); err != nil {
		panic(err)
	}

	for update := range updates {
		checkResult, err := tgBotMgr.CheckShouldHandle(&update)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "parse update error", "error": err.Error()}).Send()
		} else if !checkResult.ShouldHandle() {
			log.Info().Fields(map[string]interface{}{"action": "got unknown update", "updateId": update.UpdateID}).Send()
		} else {
			log.Info().Fields(map[string]interface{}{"action": "get should handle update", "updateId": update.UpdateID}).Send()
			tgBotMgr.HandleTGUpdate(&update, checkResult)
		}

	}

}
