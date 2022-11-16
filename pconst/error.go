package pconst

import he "github.com/tristan-club/kit/error"

func init() {
	he.InjectCodeMessage(errMap)
}

const (
	ServerError           = 500
	CodeNotSupportChainId = 40001
	CodeUserNotConfirm    = 40101
	CodeNeedPrivateChat   = 40301

	CodeRequestBotError         = 50201
	CodeBotSendMsgError         = 30001
	CodeWalletRequestError      = 30002
	CodeUserNotInit             = 30003
	CodeMarshalError            = 30004
	CodeRedisError              = 30005
	CodeInvalidCmd              = 30006
	CodeUnknownCmd              = 30007
	CodeUnknownCallbackHandler  = 30009
	CodeInvalidUserState        = 30008
	CodeAmountParamInvalid      = 30010
	CodeInvalidPayload          = 30011
	CodeEnvelopeTypeInvalid     = 30012
	CodeAssetParamInvalid       = 30013
	CodeCmdNeedGroupChat        = 30014
	CodeInsufficientBalance     = 30015
	CodeAddressParamInvalid     = 30016
	CodeInvalidChoice           = 30017
	CodeInvalidType             = 30018
	CodeInvalidQuantity         = 30019
	CodePinCodeLengthInvalid    = 30020
	CodeSamePinCode             = 30021
	CodeBridgeAmountMustInScope = 30022
	CodeBankLackBalance         = 30023
	CodeNoBankBalance           = 30024
	CodeBridgeTransactionError  = 30025
	CodeBridgeError             = 30026
	CodeCallTStoreError         = 30027
	CodeParamInvalid            = 30028
	CodeConfirmTimeOut          = 30029
	CodePinCodeInputInvalid     = 30030
	CodeGetDefaultBotFailed     = 30031
	CodePermissionRefused       = 30032
	CodeCATNotConfigure         = 30033
)

var errMap = map[int]string{
	CodeNotSupportChainId: "this chain id not supported",

	CodeRequestBotError: "an error occurred while requesting the bot service",

	CodeUserNotConfirm: "user not confirm",

	ServerError:                 "Ops, Somethings Wrong. ",
	CodeNeedPrivateChat:         "This command only works in private chat.",
	CodeCmdNeedGroupChat:        "This command only works in group chat.",
	CodeUserNotInit:             "You have not created an account, please forward to the bot and click `Start` first.",
	CodeBotSendMsgError:         "an error occur while bot sending msg",
	CodeWalletRequestError:      "an error occur while request MetaWallet service",
	CodeMarshalError:            "json resolve error",
	CodeRedisError:              "request redis error",
	CodeInvalidCmd:              "Invalid command",
	CodeUnknownCmd:              "Unknown command",
	CodeUnknownCallbackHandler:  "unknown callback chandler",
	CodeInvalidUserState:        "Ops, it looks like your user status is out of date or invalid, you can start a new command to refresh your status.",
	CodeAmountParamInvalid:      "invalid amount input",
	CodeInvalidPayload:          "cmd payload data error ",
	CodeEnvelopeTypeInvalid:     "invalid red envelope type",
	CodeAssetParamInvalid:       "asset param invalid",
	CodeInsufficientBalance:     "insufficient balance",
	CodeAddressParamInvalid:     "invalid address input",
	CodeInvalidChoice:           "command choices invalid",
	CodeInvalidType:             "invalid type input",
	CodeInvalidQuantity:         "invalid quantity input",
	CodePinCodeLengthInvalid:    "pin code should be at least 6 characters",
	CodeSamePinCode:             "new pin code same with old pin code",
	CodeBridgeAmountMustInScope: "amount input must in given range",
	CodeBankLackBalance:         "The cross-chain bridge asset pool does not have enough balance, please try later",
	CodeNoBankBalance:           "The cross-chain bridge asset pool does not have enough balance, please try later",
	CodeBridgeTransactionError:  "bridge transaction error",
	CodeBridgeError:             "bridge error",
	CodeCallTStoreError:         "request redis service error",
	CodeParamInvalid:            "invalid param",
	CodeConfirmTimeOut:          "confirm time out",
	CodePinCodeInputInvalid:     "get pin code failed",
	CodeGetDefaultBotFailed:     "default bot unavailable",
	CodePermissionRefused:       "permission refused",
	CodeCATNotConfigure:         "The current app does not configure CAT. ",
}
