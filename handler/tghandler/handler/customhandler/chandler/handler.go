package chandler

import (
	"github.com/tristan-club/kit/customid"
	"github.com/tristan-club/wizard/handler/tghandler/flow"
)

var handlerMap map[*customid.CustomId]flow.TGFlowHandler

func init() {
	handlerMap = map[*customid.CustomId]flow.TGFlowHandler{}
}

func GetCustomHandleList() map[*customid.CustomId]flow.TGFlowHandler {
	return handlerMap
}

func GetCustomHandler(cid *customid.CustomId) flow.TGFlowHandler {
	return handlerMap[cid]
}
