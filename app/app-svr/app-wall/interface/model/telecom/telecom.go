package telecom

import (
	"strconv"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
)

const (
	_telecomCardV1       = "5000003201100285" //电信9元卡
	_telecomCardV1Update = "5000003201100288" //电信9元升级卡
	_telecomCardV2       = "5000003201100284" //电信19元卡
	_telecomCardV2Update = "5000003201100287" //电信19元升级卡
	_telecomCardV3       = "5000003201100286" //电信39元卡
	_telecomCardV4       = "5000003701100418" //电信星卡套餐
)

type TelecomJSON struct {
	FlowpackageID      int            `json:"flowPackageId"`
	FlowPackageSize    int            `json:"flowPackageSize"`
	FlowPackageType    int            `json:"flowPackageType"`
	TrafficAttribution int            `json:"trafficAttribution"`
	BeginTime          string         `json:"beginTime"`
	EndTime            string         `json:"endTime"`
	IsMultiplyOrder    int            `json:"isMultiplyOrder"`
	SettlementType     int            `json:"settlementType"`
	Operator           int            `json:"operator"`
	OrderStatus        int            `json:"orderStatus"`
	RemainedRebindNum  int            `json:"remainedRebindNum"`
	MaxbindNum         int            `json:"maxBindNum"`
	OrderID            string         `json:"orderId"`
	SignNo             string         `json:"signNo"`
	AccessToken        string         `json:"accessToken"`
	PhoneID            string         `json:"phoneId"`
	IsRepeatOrder      int            `json:"isRepeatOrder"`
	PayStatus          int            `json:"payStatus"`
	PayTime            string         `json:"payTime"`
	PayChannel         int            `json:"payChannel"`
	SignStatus         string         `json:"signStatus "`
	RefundStatus       int            `json:"refundStatus"`
	PayResult          *PayResultJSON `json:"payResult,omitempty"`
}

type PayResultJSON struct {
	IsRepeatOrder int `json:"isRepeatOrder"`
	RefundStatus  int `json:"refundStatus"`
	PayStatus     int `json:"payStatus"`
	PayChannel    int `json:"payChannel"`
}

type TelecomOrderJson struct {
	RequestNo  string       `json:"requestNo"`
	ResultType int          `json:"resultType"`
	Detail     *TelecomJSON `json:"detail"`
}

type TelecomRechargeJson struct {
	RequestNo  string        `json:"requestNo"`
	ResultType int           `json:"resultType"`
	Detail     *RechargeJSON `json:"detail"`
}

type RechargeJSON struct {
	RequestNo      string `json:"requestNo"`
	FcRechargeNo   string `json:"fcRechargeNo"`
	RechargeStatus int    `json:"rechargeStatus"`
	OrderTotalSize int    `json:"orderTotalSize"`
	FlowBalance    int    `json:"flowBalance"`
}

type OrderInfo struct {
	PhoneID       int        `json:"phone"`
	OrderID       int64      `json:"orderid"`
	OrderState    int        `json:"order_status"`
	IsRepeatorder int        `json:"isrepeatorder"`
	SignNo        string     `json:"sign_no"`
	Begintime     xtime.Time `json:"begintime"`
	Endtime       xtime.Time `json:"endtime"`
}

type Pay struct {
	OrderID   int64  `json:"orderid"`
	RequestNo int64  `json:"requestno,omitempty"`
	PayURL    string `json:"pay_url,omitempty"`
}

type SucOrder struct {
	FlowPackageID        string `json:"flowPackageId,omitempty"`
	Domain               string `json:"domain"`
	Port                 string `json:"port,omitempty"`
	PortInt              int    `json:"portInt"`
	KeyEffectiveDuration int    `json:"keyEffectiveDuration"`
	OrderKey             string `json:"orderKey"`
	FlowBalance          int    `json:"flowBalance"`
	FlowPackageSize      int    `json:"flowPackageSize"`
	AccessToken          string `json:"accessToken"`
	OrderIDStr           string `json:"orderId,omitempty"`
	OrderID              int64  `json:"orderid"`
}

type OrderFlow struct {
	FlowBalance int `json:"flowBalance"`
}

type PhoneConsent struct {
	Consent int `json:"consent"`
}

type TelecomMessageJSON struct {
	PhoneID       string `json:"phoneId"`
	ResultType    int    `json:"resultType"`
	ResultMessage string `json:"resultMsg"`
}

type OrderState struct {
	FlowBalance   int        `json:"flowBalance,omitempty"`
	FlowSize      int        `json:"flow_size"`
	OrderState    int        `json:"order_state"`
	Endtime       xtime.Time `json:"endtime,omitempty"`
	IsRepeatorder int        `json:"is_repeatorder"`
}

type OrderPhoneState struct {
	FlowPackageID int    `json:"flowPackageId"`
	FlowSize      int    `json:"flowPackageSize"`
	OrderState    int    `json:"orderStatus"`
	PhoneStr      string `json:"phoneId"`
}

type CardOrder struct {
	Phone       int        `json:"phone,omitempty"`
	Nbr         string     `json:"nbr,omitempty"`
	Action      string     `json:"action,omitempty"`
	AppKey      string     `json:"appkey,omitempty"`
	ProductType int        `json:"product_type,omitempty"`
	OrderState  int        `json:"order_state,omitempty"`
	StartTime   xtime.Time `json:"start_time,omitempty"`
	EndTime     xtime.Time `json:"end_time,omitempty"`
	Spid        int        `json:"spid,omitempty"`
	Desc        string     `json:"desc,omitempty"`
}

type CardOrderBizJson struct {
	Phone      string `json:"mobile,omitempty"`
	Nbr        string `json:"nbr,omitempty"`
	StartTime  string `json:"startDate,omitempty"`
	EndTime    string `json:"endDate,omitempty"`
	CreateTime string `json:"createDate,omitempty"`
	Action     string `json:"action,omitempty"`
	AppKey     string `json:"appkey,omitempty"`
}

type CardHeadJson struct {
	SysCode       string `json:"sysCode,omitempty"`
	TransactionID string `json:"transactionId,omitempty"`
	ReqTime       string `json:"reqTime,omitempty"`
	Method        string `json:"method,omitempty"`
	Version       int    `json:"version,omitempty"`
	Attach        string `json:"attach,omitempty"`
	Sign          string `json:"sign,omitempty"`
}

type CardOrderJson struct {
	Head *CardHeadJson     `json:"head,omitempty"`
	Biz  *CardOrderBizJson `json:"biz,omitempty"`
}

type CardVipLog struct {
	Phone     int        `json:"phone,omitempty"`
	State     int8       `json:"state"`
	RequestNo int64      `json:"request_no,omitempty"`
	Ptype     int8       `json:"-"`
	Ctime     xtime.Time `json:"-"`
}

type CardAuth struct {
	Result     bool          `json:"result"`
	ComboNbr   string        `json:"comboNbr,omitempty"`
	ComboNbrs  []string      `json:"comboNbrs,omitempty"`
	CombosInfo []*CombosInfo `json:"combosInfo,omitempty"`
}

type CombosInfo struct {
	Nbr       string `json:"nbr,omitempty"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

type UserActiveParam struct {
	Build    int64  `form:"build"`
	Platform string `form:"platform"`
	Auto     bool   `form:"auto"`
	Usermob  string `form:"usermob" validate:"required"`
	Captcha  string `form:"captcha"`
	Mid      int64  `form:"-"`
	Buvid    string `form:"-"`
	IP       string `form:"-"`
}

type ActiveState struct {
	ProductID   string `json:"product_id"`
	TfType      int    `json:"tf_type"`
	TfWay       string `json:"tf_way"`
	ProductDesc string `json:"product_desc"`
	ProductTag  string `json:"product_tag"`
	ProductType int    `json:"product_type"`
}

func (s *TelecomJSON) TelecomJSONChange() {
	if s.PayResult != nil {
		s.IsRepeatOrder = s.PayResult.IsRepeatOrder
		s.RefundStatus = s.PayResult.RefundStatus
		s.PayStatus = s.PayResult.PayStatus
		s.PayChannel = s.PayResult.PayChannel
	}
}

func (t *OrderInfo) OrderInfoJSONChange(tjson *TelecomJSON) {
	t.PhoneID, _ = strconv.Atoi(tjson.PhoneID)
	t.OrderID, _ = strconv.ParseInt(tjson.OrderID, 10, 64)
	t.OrderState = tjson.OrderStatus
	t.IsRepeatorder = tjson.IsRepeatOrder
	t.SignNo = tjson.SignNo
	t.Begintime = timeStrToInt(tjson.BeginTime)
	t.Endtime = timeStrToInt(tjson.EndTime)
	t.TelecomChange()
}

// timeStrToInt
func timeStrToInt(timeStr string) (timeInt xtime.Time) {
	var err error
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(timeLayout, timeStr, loc)
	if err = timeInt.Scan(theTime); err != nil {
		log.Error("timeInt.Scan error(%v)", err)
	}
	return
}

// TelecomChange
func (t *OrderInfo) TelecomChange() {
	if t.Begintime.Time().IsZero() {
		t.Begintime = 0
	}
	if t.Endtime.Time().IsZero() {
		t.Endtime = 0
	}
}

// CardOrderChange change
func (t *CardOrder) CardOrderChange() {
	switch t.Nbr {
	case _telecomCardV1, _telecomCardV1Update:
		t.ProductType = 1
	case _telecomCardV2, _telecomCardV2Update:
		t.ProductType = 2
	case _telecomCardV3:
		t.ProductType = 3
	}
	t.Spid, _ = strconv.Atoi(t.Nbr)
	if t.StartTime.Time().IsZero() {
		t.StartTime = 0
	}
	if t.EndTime.Time().IsZero() {
		t.EndTime = 0
	}
}

func (t *CardOrder) CardOrderShow() (c *CardOrder) {
	c = &CardOrder{
		ProductType: t.ProductType,
		OrderState:  t.OrderState,
		StartTime:   t.StartTime,
		EndTime:     t.EndTime,
		Spid:        t.Spid,
	}
	return
}

func (t *CardOrder) CardAuthChange(a *CardAuth) {
	if a == nil || !a.Result || len(a.CombosInfo) == 0 {
		return
	}
	for _, info := range a.CombosInfo {
		if a.ComboNbr != info.Nbr {
			continue
		}
		switch info.Nbr {
		case _telecomCardV1, _telecomCardV1Update:
			t.ProductType = 1
			t.Desc = "真香卡"
		case _telecomCardV2, _telecomCardV2Update:
			t.ProductType = 2
			t.Desc = "真实卡"
		case _telecomCardV3:
			t.ProductType = 3
			t.Desc = "真爱卡"
		case _telecomCardV4:
			t.ProductType = 4
			t.Desc = "星卡"
		default:
			continue
		}
		t.Spid, _ = strconv.Atoi(info.Nbr)
		t.StartTime = cardTimeStrToInt(info.StartTime)
		t.EndTime = cardTimeStrToInt(info.EndTime)
		t.OrderState = 2
		break
	}
}

func cardTimeStrToInt(timeStr string) (timeInt xtime.Time) {
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(timeLayout, timeStr, loc)
	if err := timeInt.Scan(theTime); err != nil {
		log.Error("timeInt.Scan error(%v)", err)
	}
	return
}
