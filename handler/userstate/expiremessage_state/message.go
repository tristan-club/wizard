package expiremessage_state

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/bot-wizard/handler/userstate"
	"github.com/tristan-club/bot-wizard/pkg/log"
	"github.com/tristan-club/bot-wizard/pkg/util"
)

const (
	expireMessageIdListKey = "EXPIRE_MESSAGE_ID_LIST"
)

type ExpireMessage struct {
	ChatId    int64 `json:"chat_id"`
	MessageId int   `json:"message_id"`
}

func BatchAddExpireMessage(userId string, expireMessageList []*tgbotapi.Message) {
	newExpireMessageList := make([]ExpireMessage, 0)
	for _, message := range expireMessageList {
		if message == nil || message.Chat == nil {
			log.Error().Fields(map[string]interface{}{"action": "invalid message", "msg": expireMessageList}).Send()
			return
		}
		newExpireMessageList = append(newExpireMessageList, ExpireMessage{ChatId: message.Chat.ID, MessageId: message.MessageID})
	}

	param, herr := userstate.GetParam(userId, expireMessageIdListKey)
	if herr != nil {
		log.Warn().Fields(map[string]interface{}{"action": "get expire message", "error": herr.Error()}).Send()
		return
	}
	if !util.IsNil(param) {
		existExpireMessageList, ok := param.([]ExpireMessage)
		if !ok {
			us, _ := userstate.GetState(userId, nil)
			log.Warn().Fields(map[string]interface{}{"action": "invalid expire message id list", "user id": userId, "user state": us}).Send()
			return
		}

		newExpireMessageList = append(newExpireMessageList, existExpireMessageList...)

	}
	userstate.SetParam(userId, expireMessageIdListKey, newExpireMessageList)
}

func AddExpireMessage(userId string, message *tgbotapi.Message) {

	BatchAddExpireMessage(userId, []*tgbotapi.Message{message})
}

func GetExpireMessage(userId string) (resp []ExpireMessage) {
	resp = make([]ExpireMessage, 0)
	param, herr := userstate.GetParam(userId, expireMessageIdListKey)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "get expire message", "error": herr.Error()}).Send()
		return
	}

	if !util.IsNil(param) {

		b, err := json.Marshal(param)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "expire message marshal", "error": err.Error()}).Send()
			return
		}
		if err := json.Unmarshal(b, &resp); err != nil {
			log.Error().Fields(map[string]interface{}{"action": "expire  message marhshal", "error": err.Error()}).Send()
			return
		}

		return
	}
	return
}

func ClearExpireMessage(userId string) {
	userstate.SetParam(userId, expireMessageIdListKey, nil)
}
