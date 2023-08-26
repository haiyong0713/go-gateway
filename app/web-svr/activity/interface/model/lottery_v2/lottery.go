package lottery

import (
	xtime "go-common/library/time"
)

const (
	silenceForbid = 1
	vipCheck      = 1
	monthVip      = 2
	yearVip       = 3
	telValid      = 1
	identifyValid = 2
	// UsedTimesKey 使用次数
	UsedTimesKey = "used"
	// AddTimesKey 增加的次数
	AddTimesKey = "add"
	// DailyAddType 每日过期类型
	DailyAddType = 1
	// HightTypeBuyVip 购买大会员优先级
	HightTypeBuyVip = 1
	// HightTypeArchive 投稿优先级
	HightTypeArchive = 2
	// MustWinRate 必中奖
	MustWinRate = 1
	// CoinConsume 抽奖消耗
	CoinConsume = "抽奖消耗"
	// IsInternal 是内部抽奖
	IsInternal = 1
	// MsgTypeCard 通知卡类型 type = 10
	MsgTypeCard = 10
	// IsRookie ...
	IsRookie = 1
)

// Page represents the standard page structure
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// WinList ///
type WinList struct {
	Name string `json:"name"`
	*GiftMid
}

// RealWinList ...
type RealWinList struct {
	List []*MidWinList `json:"list"`
}

// RecordReply ...
type RecordReply struct {
	List         []*RecordDetail `json:"list"`
	Page         *Page           `json:"page"`
	IsAddAddress bool            `json:"is_add_address"`
}

// FrontEndParams ...
type FrontEndParams struct {
	// Ip ip
	IP string
	// DeviceId ...
	DeviceID string
	// Ua ...
	Ua string
	// API ...
	API string
	// Referer ...
	Referer string
}

// Lottery lottery
type Lottery struct {
	ID         int64      `json:"id"`
	LotteryID  string     `json:"lottery_id"`
	Name       string     `json:"name"`
	IsInternal int        `json:"is_internal"`
	Stime      xtime.Time `json:"stime"`
	Etime      xtime.Time `json:"etime"`
	Ctime      xtime.Time `json:"ctime"`
	Mtime      xtime.Time `json:"mtime"`
	Type       int        `json:"type"`
	State      int        `json:"state"`
	Author     string     `json:"author"`
}

// Info detail
type Info struct {
	ID           int64      `json:"id"`
	Sid          string     `json:"sid"`
	Level        int        `json:"level"`
	RegTimeStime int64      `json:"regtime_stime"`
	RegTimeEtime int64      `json:"regtime_etime"`
	VipCheck     int        `json:"vip_check"`
	AccountCheck int        `json:"account_check"`
	Coin         int        `json:"coin"`
	FsIP         int        `json:"fs_ip"`
	GiftRate     int64      `json:"gift_rate"`
	HighType     int        `json:"high_type"`
	HighRate     int64      `json:"high_rate"`
	SenderID     int64      `json:"sender_id"`
	ActivityLink string     `json:"activity_link"`
	FigureScore  int64      `json:"figure_score"`
	SpyScore     int64      `json:"spy_score"`
	Ctime        xtime.Time `json:"ctime"`
	Mtime        xtime.Time `json:"mtime"`
	State        int        `json:"state"`
}

// LetterParam for private msg param
type LetterParam struct {
	RecverIDs  []uint64 `json:"recver_ids"`       //多人消息，列表型，限定每次客户端发送<=100
	SenderUID  uint64   `json:"sender_uid"`       //官号uid：发送方uid
	MsgKey     uint64   `json:"msg_key"`          //消息唯一标识
	MsgType    int32    `json:"msg_type"`         //文本类型 type = 1
	Content    string   `json:"content"`          //{"content":"test" //文本内容}
	NotifyCode string   `json:"notify_code"`      //通知码
	Params     string   `json:"params,omitempty"` //逗号分隔，通知卡片内容的可配置参数
	JumpURL    string   `json:"jump_url"`         //通知卡片跳转链接
	JumpURL2   string   `json:"jump_url2"`        //通知卡片跳转链接
	Title      string   `json:"title"`
	Text       string   `json:"text"`
	JumpText   string   `json:"jump_text"`
	JumpText2  string   `json:"jump_text2"`
}
