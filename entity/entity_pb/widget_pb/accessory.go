package widget_pb

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/tristan-club/kit/log"
)

func (c *TaskTgJoinGroup) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *TaskTgJoinGroup) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}

func (c *TaskTgInviteGroup) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *TaskTgInviteGroup) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}

func (c *TaskDiscordJoinServer) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *TaskDiscordJoinServer) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}

func (c *TaskDiscordInviteServer) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *TaskDiscordInviteServer) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}

func (c *TaskTwitterFollow) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *TaskTwitterFollow) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}

func (c *TaskTwitterRetweet) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *TaskTwitterRetweet) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}

func (c *TaskBindWallet) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *TaskBindWallet) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}

func (c *SendGroup) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *SendGroup) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}

func (c *AwardList) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

func (c *AwardList) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	if err != nil {
		log.Error().Msgf("json marshal error:%s", err.Error())
	}
	return string(b), err
}
