package tcontext

import (
	"github.com/tristan-club/bot-wizard/handler/userstate"
	"time"
)

var deadlineMsg = map[int]int64{}

func (ctx *Context) SetDeadlineMsg(chatId int64, messageId int, deadline time.Duration) {
	deadlineMsg[messageId] = chatId
	go func() {
		time.Sleep(deadline)
		if cid, ok := deadlineMsg[messageId]; ok && cid > 0 {
			ctx.DeleteMessage(cid, messageId)
			userstate.ResetState(ctx.Requester.RequesterOpenId)
		}
	}()
}

func (ctx *Context) RemoveDeadlineMsg(messageId int) {
	delete(deadlineMsg, messageId)
}
