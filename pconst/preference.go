package pconst

import "time"

const (
	COMMON_KEYBOARD_DEADLINE = 60 * time.Second
	COMMON_MSG_DEADLINE      = 30 * time.Second
	ForwardPrivateDeadline   = 15 * time.Second
	GroupMentionDeadline     = 15 * time.Second
)
