package model

// LetterParam for private msg param
type LetterParam struct {
	RecverIDs  []uint64 `json:"recver_ids"`       //多人消息，列表型，限定每次客户端发送<=100
	SenderUID  uint64   `json:"sender_uid"`       //官号uid：发送方uid
	MsgKey     uint64   `json:"msg_key"`          //消息唯一标识
	MsgType    int32    `json:"msg_type"`         //文本类型 type = 1
	Content    string   `json:"content"`          //{"content":"test" //文本内容}
	NotifyCode string   `json:"notify_code"`      //通知码
	Params     string   `json:"params,omitempty"` //逗号分隔，通知卡片内容的可配置参数
	JumpUri    string   `json:"jump_uri"`         //通知卡片跳转链接
	Title      string   `json:"title"`
	Text       string   `json:"text"`
	JumpText   string   `json:"jump_text"`
	JumpUri2   string   `json:"jump_uri_2"` //通知卡片跳转链接，带上此参数则以参数里的url为准，否则以后台申请破冰码时录入的url为准
}
