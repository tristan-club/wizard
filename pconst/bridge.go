package pconst

const (
	BridgeStatusCreate = iota + 1
	BridgeStatusFromTxPending
	BridgeStatusFromTxSuccess
	BridgeStatusFromTxFailed
	BridgeStatusToTxPending
	BridgeStatusToTxSuccess
	BridgeStatusToTxFailed
	BridgeStatusOtherFailed = 10
)

const BridgeGasAmount = "0.003"
