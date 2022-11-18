package pconst

const (
	CODE_ENVELOPE_SOLD_OUT     = 1202
	CODE_ENVELOPE_OPENED       = 1203
	CODE_ENVELOP_STATE_INVALID = 1204
)

const (
	EnvelopStatusCreated         = iota + 1 // 红包创建成功
	EnvelopStatusRechargeSuccess            // 红包创建者向红包账户充值成功,进入抢的过程
	EnvelopStatusRechargeFailed             // 红包创建者向红包账户充值失败
	//EnvelopStatusSoldOut                    // 红包在过期时间内被抢完
	EnvelopStatusExpired // 红包已过期且未被抢完
)

const (
	EnvelopeStorePrefix        = "ENVELOPE_"
	EnvelopeStorePathMsgId     = "ENVELOPE_MSG_ID"
	EnvelopeStorePathChannelId = "ENVELOPE_CH_ID"
)

const (
	EnvelopeTypeOrdinary = 1
	EnvelopeTypeTask     = 2
)
