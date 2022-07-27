package main

import (
	"github.com/tristan-club/wizard/cmd"
	"github.com/tristan-club/wizard/manager/tgmgr"
	"os"
	"time"
)

func main() {

	time.Local = time.FixedZone("UTC", 0)

	tgBotMgr, err := tgmgr.NewTGMgr(os.Getenv("CONTROLLER_SERVICE"), os.Getenv("TSTORE_SERVICE"), cmd.GetCmdList())
	if err != nil {
		panic(err)
	}

	if err := tgBotMgr.ListenTGUpdate(os.Getenv("TG_BOT_TOKEN"), "", ""); err != nil {
		panic(err)
	}
}
