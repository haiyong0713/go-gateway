package family

const (
	// identity
	IdentityParent = "parent"
	IdentityChild  = "child"
	IdentityNormal = "" //普通用户
	// max_bind
	MaxBind = 3
	// teen_action
	ParentActionOpen  = "open"
	ParentActionClose = "close"
	// ticket alphabet
	TicketAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	TicketLength   = 13
	// relation_type
	RelTypeNormal = 0
	RelTypeParent = 1
	RelTypeChild  = 2
	// timelock push_time
	TLPushTime = 3 //离触发时间还剩X分钟
	// default daily_duration
	DefaultDailyDuration = 40
)
