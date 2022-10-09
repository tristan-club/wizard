package tcontext

import (
	"time"
)

var deadlineMsg = map[int]int64{}

func (ctx *Context) SetDeadlineMsg(chatId int64, messageId int, deadline time.Duration) {
	deadlineMsg[messageId] = chatId
	go func() {
		time.Sleep(deadline)
		if cid, ok := deadlineMsg[messageId]; ok && cid != 0 {
			ctx.DeleteMessage(cid, messageId)
			delete(deadlineMsg, messageId)
		}
	}()
}

func (ctx *Context) RemoveDeadlineMsg(messageId int) {
	delete(deadlineMsg, messageId)
}
