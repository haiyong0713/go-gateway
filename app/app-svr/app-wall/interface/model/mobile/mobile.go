package mobile

import (
	"encoding/xml"

	xtime "go-common/library/time"
)

type OrderXML struct {
	XMLName xml.Name `xml:"SyncFlowPkgOrderReq"`
	*MobileXML
}

type FlowXML struct {
	XMLName xml.Name `xml:"SyncFlowPkgLeftQuotaReq"`
	*MobileXML
}

type MobileXML struct {
	Orderid        string `xml:"OrderID"`
	Userpseudocode string `xml:"UserPseudoCode"`
	Channelseqid   string `xml:"ChannelSeqId"`
	Price          string `xml:"Price"`
	Actiontime     string `xml:"ActionTime"`
	Actionid       string `xml:"ActionID"`
	Effectivetime  string `xml:"EffectiveTime"`
	Expiretime     string `xml:"ExpireTime"`
	Channelid      string `xml:"ChannelId"`
	Productid      string `xml:"ProductId"`
	Ordertype      string `xml:"OrderType"`
	Threshold      string `xml:"Threshold"`
	Resulttime     string `xml:"ResultTime"`
}

type Mobile struct {
	Orderid        string     `json:"-"`
	Userpseudocode string     `json:"-"`
	Channelseqid   string     `json:"-"`
	Price          int        `json:"-"`
	Actionid       int        `json:"actionid"`
	Effectivetime  xtime.Time `json:"starttime,omitempty"`
	Expiretime     xtime.Time `json:"endtime,omitempty"`
	Channelid      string     `json:"-"`
	Productid      string     `json:"productid,omitempty"`
	Ordertype      int        `json:"-"`
	Threshold      int        `json:"flow"`
	Resulttime     xtime.Time `json:"-"`
	MobileType     int        `json:"orderstatus,omitempty"`
	ProductType    int        `json:"product_type,omitempty"`
	Desc           string     `json:"desc,omitempty"`
	ProductTag     string     `json:"product_tag,omitempty"`
	ProductID      string     `json:"-"`
}

type MobileIP struct {
	IPStartUint uint32 `json:"-"`
	IPEndUint   uint32 `json:"-"`
}

type MobileUserIP struct {
	IPStr    string `json:"ip"`
	IsValide bool   `json:"is_valide"`
}

type Msg struct {
	Xmlns   string `xml:"xmlns,attr"`
	MsgType string `xml:"MsgType"`
	Version string `xml:"Version"`
	HRet    string `xml:"hRet"`
}

type OrderMsgXML struct {
	XMLName xml.Name `xml:"SyncFlowPkgOrderResp"`
	*Msg
}

type FlowMsgXML struct {
	XMLName xml.Name `xml:"SyncFlowPkgLeftQuotaResp"`
	*Msg
}

type UserActiveParam struct {
	Build    int64  `form:"build"`
	Platform string `form:"platform"`
	Usermob  string `form:"usermob" validate:"required"`
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
