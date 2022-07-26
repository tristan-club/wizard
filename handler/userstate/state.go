package userstate

import (
	"encoding/json"
	"fmt"
	he "github.com/tristan-club/bot-wizard/pkg/error"
	"github.com/tristan-club/bot-wizard/pkg/log"
	"github.com/tristan-club/bot-wizard/pkg/tstore"
	"github.com/tristan-club/bot-wizard/pkg/util"
)

const (
	ServiceNone = "none"
	CmdNone     = "none"
	StateNone   = 0
)

const (
	PathUserState = "USER_STATE_V2"
)

type UserState struct {
	OpenId         string                 `json:"open_id"`
	UserNo         string                 `json:"user_no"`
	CurrentState   int                    `json:"current_state"`
	CurrentCommand string                 `json:"current_command"`
	CurrentService string                 `json:"current_service"`
	DefaultAddress string                 `json:"default_address"`
	Payload        map[string]interface{} `json:"payload"`
}

func SetState(openId string, state int, currentCmd, currentService string, payload map[string]interface{}) he.Error {
	us, herr := GetState(openId, nil)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "get user state error", "error": herr.Error()}).Send()
		return herr
	}

	if state != 0 {
		us.CurrentState = state
	}

	if currentCmd != "" {
		us.CurrentCommand = currentCmd
	}

	if currentService != "" {
		us.CurrentService = currentService
	}

	if len(payload) != 0 {
		us.Payload = payload
	}

	usByte, err := json.Marshal(us)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "marshal user state", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeMarshalError, "", err)
	}
	if err := tstore.PBSaveString(openId, PathUserState, string(usByte)); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "tstore save", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeCallTStoreError, "", err)
	}

	log.Info().Fields(map[string]interface{}{"action": "success update user state", "OpenId": openId, "state": state,
		"service": currentService}).Send()

	return nil
}

func InitState(openId, cmd, userId, defaultAddress string) he.Error {

	userState := &UserState{
		OpenId:         openId,
		UserNo:         userId,
		CurrentState:   StateNone,
		CurrentCommand: cmd,
		CurrentService: ServiceNone,
		DefaultAddress: defaultAddress,
		Payload:        map[string]interface{}{},
	}

	usByte, err := json.Marshal(userState)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "marshal user state", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeMarshalError, "", err)
	}
	if err := tstore.PBSaveString(openId, PathUserState, string(usByte)); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "tstore save", "error": err.Error()}).Send()
		return he.NewServerError(he.CodeMarshalError, "", err)
	}

	return nil

}

func ResetState(openId string) he.Error {
	log.Info().Fields(map[string]interface{}{"action": "reset user state", "OpenId": openId}).Send()
	return SetState(openId, StateNone, CmdNone, ServiceNone, nil)
}

func GetState(openId string, out interface{}) (*UserState, he.Error) {
	userState := &UserState{
		OpenId:       openId,
		CurrentState: StateNone,
		Payload:      map[string]interface{}{},
	}

	userStateSaveStr, err := tstore.PBGetStr(openId, PathUserState)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "tstore get", "error": err.Error()}).Send()
		return nil, he.NewServerError(he.CodeCallTStoreError, "", err)
	}

	if userStateSaveStr == "" {
		return userState, nil
	}

	if err := json.Unmarshal([]byte(userStateSaveStr), &userState); err != nil {
		log.Error().Fields(map[string]interface{}{"action": "marshal user state", "error": err.Error()}).Send()
		return nil, he.NewServerError(he.CodeMarshalError, "", err)
	}

	if !util.IsNil(out) && len(userState.Payload) != 0 {
		b, err := json.Marshal(userState.Payload)
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "marshal payload", "error": err.Error()}).Send()
			return nil, he.NewServerError(he.CodeMarshalError, "", err)
		}
		if err := json.Unmarshal(b, &out); err != nil {
			log.Error().Fields(map[string]interface{}{"action": "unmarshal payload", "error": err.Error()}).Send()
			return nil, he.NewServerError(he.CodeMarshalError, "", err)
		}
	}
	if userState.Payload == nil {
		userState.Payload = map[string]interface{}{}
	}

	log.Info().Fields(map[string]interface{}{"action": "get user state success", "OpenId": openId, "state": userState.CurrentState,
		"service": userState.CurrentService, "cmd": userState.CurrentCommand}).Send()
	return userState, nil
}

func SetParam(openId string, key string, value interface{}) {
	us, herr := GetState(openId, nil)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "get user state", "error": herr.Error()}).Send()
		return
	}
	us.Payload[key] = value
	herr = SetState(openId, 0, "", "", us.Payload)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "save payload", "error": herr.Error()}).Send()
	}
}

func BatchSaveParam(openId string, params map[string]interface{}) {
	us, herr := GetState(openId, nil)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "get user state", "error": herr.Error()}).Send()
		return
	}
	for k, v := range params {
		us.Payload[k] = v
	}
	herr = SetState(openId, 0, "", "", us.Payload)
	if herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "save payload", "error": herr.Error()}).Send()
	}
}

func GetParam(openId, key string) (interface{}, he.Error) {
	us, herr := GetState(openId, nil)
	if herr != nil {
		return nil, herr
	}
	return us.Payload[key], nil
}

func MustString(openId string, key string) (string, he.Error) {
	us, _ := GetState(openId, nil)
	if resp, ok := us.Payload[key].(string); !ok || resp == "" {
		return "", he.NewServerError(he.CodeInvalidPayload, "", fmt.Errorf("need string get %s", util.FastMarshal(us.Payload[key])))
	} else {
		return resp, nil
	}
}

func MustInt64(openId string, key string) (int64, he.Error) {
	us, _ := GetState(openId, nil)
	param := us.Payload[key]
	switch param.(type) {
	case int64:
		return param.(int64), nil
	case float64:
		return int64(param.(float64)), nil
	default:
		return 0, he.NewServerError(he.CodeInvalidPayload, "", fmt.Errorf("need int get %s", util.FastMarshal(us.Payload[key])))
	}
}

func MustUInt64(openId string, key string) (uint64, he.Error) {
	resp, herr := MustInt64(openId, key)
	if herr != nil {
		return 0, herr
	} else if resp == 0 {
		return 0, he.NewServerError(he.CodeInvalidPayload, "", fmt.Errorf("got 0 for uint64, OpenId %s, key %s", openId, key))
	}
	return uint64(resp), nil
}
