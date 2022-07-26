package pconst

import "time"

const (
	QueryTxPending = time.Hour * 1
)

const (
	TxStatePending = iota + 1
	TxStateSuccess
	TxStateFailed
)
