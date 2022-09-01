package presetnode

import (
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/config"
	"github.com/tristan-club/wizard/handler/text"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain/presetnode/prechecker"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/handler/userstate/expiremessage_state"
	"github.com/tristan-club/wizard/pconst"
	"strings"
)

var EnterPinCodeHandler = chain.NewNode(AskForPinCode, prechecker.MustBeMessage, EnterPinCode)

type EnterPinCodeParam struct {
	Content             string `json:"content"`
	ParamKey            string `json:"param_key"`
	UseTargetContentKey string `json:"use_target_content_key"`
	UserMarkdown        bool   `json:"user_markdown"`
}

func AskForPinCode(ctx *tcontext.Context, node *chain.Node) error {

	var param = &EnterPinCodeParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	content := param.Content
	if content == "" {
		if param.UseTargetContentKey != "" {
			targetContent, herr := userstate.MustString(ctx.OpenId(), param.UseTargetContentKey)
			if herr != nil {
				log.Error().Fields(map[string]interface{}{"action": "get target content", "error": herr}).Send()
				return herr
			}
			if targetContent != "" {
				content = targetContent
			}
		}
	}

	if content == "" {
		content = text.EnterPinCode
	}
	if param.UserMarkdown {
		content = strings.ReplaceAll(content, ".", "\\.")
	}

	if msg, herr := ctx.Send(ctx.U.SentFrom().ID, content, nil, param.UserMarkdown, false); herr != nil {
		return herr
	} else {
		if param.UseTargetContentKey == "" {
			expiremessage_state.AddExpireMessage(ctx.OpenId(), msg)
		}
	}

	return nil
}

func EnterPinCode(ctx *tcontext.Context, node *chain.Node) error {
	var param = &EnterPinCodeParam{}
	if !node.IsPayloadNil() {
		herr := node.TryGetPayload(param)
		if herr != nil {
			return herr
		}
	}
	paramKey := param.ParamKey
	if paramKey == "" {
		paramKey = "pin_code"
	}
	pinCode := ctx.U.Message.Text

	if !config.EnvIsDev() && len(pinCode) < 6 {
		return he.NewBusinessError(pconst.CodePinCodeLengthInvalid, "", nil)
	}

	if pinCode == "cancel" {
		return he.NewBusinessError(he.BusinessError, "cancel the process", nil)
	}

	userstate.SetParam(ctx.OpenId(), paramKey, pinCode)

	if herr := ctx.DeleteMessage(ctx.U.FromChat().ID, ctx.U.Message.MessageID); herr != nil {
		log.Error().Fields(map[string]interface{}{"action": "delete pin code error", "error": herr}).Send()
	}

	return nil
}
