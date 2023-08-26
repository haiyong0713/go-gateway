package app

// CometNotify消息服务 - 接口文档地址: https://info.bilibili.co/pages/viewpage.action?pageId=97934351

const (
	NOTIDY_WECHART_MESSAGE_TYPE_TEXT = 1
	NOTIDY_WECHART_MESSAGE_TYPE_CARD = 2
)

type CometNotifyReq struct {
	App     string           `json:"app"`
	Users   []*CometUserInfo `json:"users"`
	Message *CometMessageSet `json:"new_message"`
}

type CometUserInfo struct {
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`
}

type CometMessageSet struct {
	WechatMessage    *CometWeChatMessage    `json:"new_wechat,omitempty"`
	EmailMessage     *CometEmailMessage     `json:"new_email,omitempty"`
	PhoneCallMessage *CometPhoneCallMessage `json:"new_call,omitempty"`
}

type CometPhoneCallMessage struct {
	MessageType     string `json:"message_type"`
	Message         string `json:"message"`
	MessageMetadata string `json:"message_metadata"`
}

type CometEmailMessage struct {
	MessageType     string      `json:"message_type"`
	Message         string      `json:"message"`
	MessageMetadata interface{} `json:"message_metadata"`
	Template        string      `json:"template,omitempty"`
}

type CometWeChatMessage struct {
	MessageType     string      `json:"message_type"`
	Message         interface{} `json:"message"`
	MessageMetadata interface{} `json:"message_metadata"`
	Template        string      `json:"template,omitempty"`
	MessageDetail   string      `json:"message_detail,omitempty"`
}

type CometPictureMessage struct {
	MessageType     string      `json:"message_type"`
	Message         string      `json:"message"`
	MessageMetadata interface{} `json:"message_metadata"`
	MessageDetail   string      `json:"message_detail,omitempty"`
}

type CometPictureArticle struct {
	Articles []*CometPictureArticleContent `json:"articles"`
}

type CometPictureArticleContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	Picurl      string `json:"picurl"`
}

type CometEmailMetadata struct {
	Subject string `json:"subject"`
	Sender  string `json:"sender"`
}

type CometWeChatCardMetadata struct {
	Subject string `json:"subject"`
	Link    string `json:"link,omitempty"`
}

type CometPictureMetadata struct {
	WxPicture bool `json:"wx_picture"`
}
