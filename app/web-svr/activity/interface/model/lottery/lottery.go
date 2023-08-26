package lottery

import (
	xtime "go-common/library/time"
)

const (
	// MsgTypeCard 消息形式通知卡
	MsgTypeCard = 10
	// IsShow 展示
	IsShow = 1
)

const (
	// TimesAddTimesStateNone 不能领取
	TimesAddTimesStateNone = 1
	// TimesAddTimesStateWait 待领取
	TimesAddTimesStateWait = 2
	// TimesAddTimesStateAlready 已领取
	TimesAddTimesStateAlready = 3
)

// TimesInfo ...
type TimesInfo struct {
	Counter  string `json:"counter"`
	Activity string `json:"activity"`
	Count    int64  `json:"count"`
}
type Lottery struct {
	ID        int64      `json:"id"`
	LotteryID string     `json:"lottery_id"`
	Name      string     `json:"name"`
	Stime     xtime.Time `json:"stime"`
	Etime     xtime.Time `json:"etime"`
	Ctime     xtime.Time `json:"ctime"`
	Mtime     xtime.Time `json:"mtime"`
	Type      int        `json:"type"`
	State     int        `json:"state"`
}

type LotteryInfo struct {
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
	Ctime        xtime.Time `json:"ctime"`
	Mtime        xtime.Time `json:"mtime"`
	State        int        `json:"state"`
}

type LotteryTimesConfig struct {
	ID      int64  `json:"id"`
	Sid     string `json:"sid"`
	Type    int    `json:"type"`
	AddType int    `json:"add_type"`
	Times   int    `json:"times"`
	Info    string `json:"info"`
	Most    int    `json:"most"`
	State   int    `json:"state"`
}

// CountStateReply ...
type CountStateReply struct {
	State int `json:"state"`
}

type LotteryGift struct {
	ID             int64      `json:"id"`
	Sid            string     `json:"sid"`
	Ctime          xtime.Time `json:"ctime"`
	Mtime          xtime.Time `json:"mtime"`
	Name           string     `json:"name"`
	Num            int64      `json:"num"`
	Type           int        `json:"type"`
	Source         string     `json:"source"`
	ImgUrl         string     `json:"img_url"`
	TimeLimit      xtime.Time `json:"time_limit"`
	IsShow         int        `json:"is_show"`
	LeastMark      int        `json:"least_mark"`
	MessageTitle   string     `json:"message_title"`
	MessageContent string     `json:"message_content"`
	SendNum        int64      `json:"send_num"`
	Efficient      int        `json:"efficient"`
	State          int        `json:"state"`
}

// addr
type AddressInfo struct {
	ID      int64  `json:"id"`
	Type    int64  `json:"type"`
	Def     int64  `json:"def"`
	ProvID  int64  `json:"prov_id"`
	CityID  int64  `json:"city_id"`
	AreaID  int64  `json:"area_id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Addr    string `json:"addr"`
	ZipCode string `json:"zip_code"`
	Prov    string `json:"prov"`
	City    string `json:"city"`
	Area    string `json:"area"`
}

type LotteryRecordDetail struct {
	ID       int64      `json:"id"`
	Mid      int64      `json:"mid"`
	IP       int64      `json:"ip"`
	Num      int        `json:"num"`
	GiftID   int64      `json:"gift_id"`
	GiftName string     `json:"gift_name"`
	GiftType int        `json:"gift_type"`
	ImgURL   string     `json:"img_url"`
	Type     int        `json:"type"`
	Ctime    xtime.Time `json:"ctime"`
	CID      int64      `json:"cid"`
}

type LotteryRecordRes struct {
	List         []*LotteryRecordDetail `json:"list"`
	Page         *Page                  `json:"page"`
	IsAddAddress bool                   `json:"is_add_address"`
}

type LotteryTimesRes struct {
	Times int `json:"times"`
}

type LotteryAddTimes struct {
	ID    int64      `json:"id"`
	Mid   int64      `json:"mid"`
	Type  int        `json:"type"`
	Num   int        `json:"num"`
	CID   int64      `json:"cid"`
	Ctime xtime.Time `json:"ctime"`
}

type GiftList struct {
	GiftID     int64      `json:"gift_id"`
	GiftName   string     `json:"gift_name"`
	GiftImgUrl string     `json:"gift_img_url"`
	Mid        int64      `json:"mid"`
	Ctime      xtime.Time `json:"ctime"`
}

// Page represents the standard page structure
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

type WinList struct {
	Name string `json:"name"`
	*GiftList
}

type InsertRecord struct {
	Mid     int64  `json:"mid"`
	Num     int    `json:"num"`
	Type    int    `json:"type"`
	CID     int64  `json:"cid"`
	OrderNo string `json:"string"`
}

type GrantJson struct {
	Pid    int64 `json:"pid"`
	Expire int64 `json:"expire"`
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

type LottArg struct {
	Sid          int64                          `json:"sid"`
	Ip           string                         `json:"ip"`
	Mid          int64                          `json:"mid"`
	Rate         int64                          `json:"rate"`
	Lottery      *Lottery                       `json:"lottery"`
	LotteryMap   map[string]*LotteryTimesConfig `json:"lotteryMap"`
	GiftMap      map[int64]*LotteryGift         `json:"giftMap"`
	InsertRecord []*InsertRecord                `json:"insertRecord"`
	High         int                            `json:"high"`
	CanLottery   int                            `json:"canLottery"`
	Win          int64                          `json:"win"`
}

type ConsumeArg struct {
	Sid           int64                          `json:"sid"`
	Mid           int64                          `json:"mid"`
	AddMap        map[string]int                 `json:"addMap"`
	UsedMap       map[string]int                 `json:"usedMap"`
	LotteryMap    map[string]*LotteryTimesConfig `json:"lotteryMap"`
	Num           int                            `json:"num"`
	Base          int64                          `json:"base"`
	Share         int64                          `json:"share"`
	Follow        int64                          `json:"follow"`
	Other         int64                          `json:"other"`
	LikeList      []int64                        `json:"likeList"`
	BuyList       []int64                        `json:"buyList"`
	CustomizeList []int64                        `json:"customizeList"`
	OgvList       []int64                        `json:"ogv_list"`
	Fe            int64                          `json:"fe"`
	TimesLike     int64                          `json:"times_like"`
	TimesCoin     int64                          `json:"times_coin"`
}

type LotteryCard struct {
	Num  int           `json:"num"`
	Card map[int64]int `json:"card"`
}
