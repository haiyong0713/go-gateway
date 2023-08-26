package unicom

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-wall/interface/model"
)

type FreeProduct string

var (
	FlowProduct = FreeProduct("flow")
	CardProduct = FreeProduct("card")
)

type ActivateResponse struct {
	Result   string `json:"result"`
	Service  string `json:"service"`
	Function string `json:"function"`
	Errcode  string `json:"errcode"`
	Flag     string `json:"flag"`
	Product  string `json:"product"`
	Package  string `json:"package"`
}

type Unicom struct {
	Id          int        `json:"-"`
	Spid        int        `json:"spid"`
	CardType    int        `json:"cardtype"`
	TypeInt     int        `json:"type"`
	Unicomtype  int        `json:"unicomtype,omitempty"`
	Ordertypes  int        `json:"-"`
	Channelcode int        `json:"-"`
	Usermob     string     `json:"-"`
	Cpid        string     `json:"-"`
	Ordertime   xtime.Time `json:"ordertime"`
	Canceltime  xtime.Time `json:"canceltime,omitempty"`
	Endtime     xtime.Time `json:"endtime,omitempty"`
	Province    string     `json:"-"`
	Area        string     `json:"-"`
	Videoid     string     `json:"-"`
	Time        xtime.Time `json:"-"`
	Flowbyte    int        `json:"flowbyte"`
	Flowtype    int        `json:"flowtype,omitempty"`
	Desc        string     `json:"desc,omitempty"`
	ProductTag  string     `json:"product_tag,omitempty"`
	ProductType int        `json:"-"`
	TfWay       string     `json:"-"`
	TfType      int        `json:"-"`
}

type UnicomJson struct {
	Usermob     string `json:"usermob"`
	Userphone   string `json:"userphone"`
	Cpid        string `json:"cpid"`
	Spid        string `json:"spid"`
	TypeInt     string `json:"type"`
	Ordertime   string `json:"ordertime"`
	Canceltime  string `json:"canceltime"`
	Endtime     string `json:"endtime"`
	Channelcode string `json:"channelcode"`
	Province    string `json:"province"`
	Area        string `json:"area"`
	Ordertypes  string `json:"ordertype"`
	Videoid     string `json:"videoid"`
	Time        string `json:"time"`
	FlowbyteStr string `json:"flowbyte"`
	FakeID      string `json:"fakeid"`
	FakeIDMonth string `json:"fakeidmonth"`
}

type UnicomIpJson struct {
	Ipbegin   string `json:"ipbegin"`
	Ipend     string `json:"ipend"`
	Provinces string `json:"province"`
	Isopen    string `json:"isopen"`
	Opertime  string `json:"opertime"`
	Sign      string `json:"sign"`
}

type UnicomIP struct {
	Ipbegin     int    `json:"-"`
	Ipend       int    `json:"-"`
	IPStartUint uint32 `json:"-"`
	IPEndUint   uint32 `json:"-"`
}

type UnicomUserIP struct {
	IPStr    string `json:"ip"`
	IsValide bool   `json:"is_valide"`
}

type BroadbandOrder struct {
	Usermob string `json:"userid,omitempty"`
	Endtime string `json:"endtime,omitempty"`
	Channel string `json:"channel,omitempty"`
}

type UserBind struct {
	Usermob  string    `json:"usermob,omitempty"`
	Phone    int       `json:"phone"`
	Mid      int64     `json:"mid"`
	Name     string    `json:"name,omitempty"`
	State    int       `json:"state,omitempty"`
	Integral int       `json:"integral"`
	Flow     int       `json:"flow"`
	Monthly  time.Time `json:"monthly,omitempty"`
}

type UserBindV2 struct {
	Mid         int64  `json:"mid,omitempty"`
	Usermob     string `json:"-"`
	Integral    int    `json:"integral,omitempty"`
	Flow        int    `json:"flow,omitempty"`
	BindingTime string `json:"binding_time,omitempty"`
	UpdateTime  string `json:"update_time,omitempty"`
}

// UserPack is the offer for the user to choose
type UserPack struct {
	ID       int64  `json:"id"`
	Type     int    `json:"type"`
	Desc     string `json:"desc"`
	Amount   int    `json:"amount"`
	Capped   int8   `json:"capped"`
	Integral int    `json:"integral"`
	Param    string `json:"param"`
	State    int    `json:"state,omitempty"`
	Original int    `json:"original,omitempty"`
	Kind     int    `json:"kind"`
	Cover    string `json:"cover"`
	NewParam string `json:"new_param"`
}

const (
	_typeComic   = 4
	_typeTraffic = 0
)

// IsComic def.
func (v *UserPack) IsComic() bool {
	return v.Type == _typeComic
}

// IsTraffic def.
func (v *UserPack) IsTraffic() bool {
	return v.Type == _typeTraffic
}

type UserPackLimit struct {
	IsLimit int `json:"is_limit"`
	Count   int `json:"count"`
}

type UnicomUserFlow struct {
	Phone      int    `json:"phone"`
	Mid        int64  `json:"mid"`
	Integral   int    `json:"integral"`
	Flow       int    `json:"flow"`
	Outorderid string `json:"outorderid"`
	Orderid    string `json:"orderid"`
	Desc       string `json:"desc"`
}

type UserPackLog struct {
	Phone     int    `json:"phone,omitempty"`
	Usermob   string `json:"usermob,omitempty"`
	Mid       int64  `json:"mid,omitempty"`
	RequestNo string `json:"request_no,omitempty"`
	Type      int    `json:"pack_type"`
	Desc      string `json:"-"`
	UserDesc  string `json:"pack_desc,omitempty"`
	Integral  int    `json:"integral,omitempty"`
}

type UserLog struct {
	Phone    int    `json:"phone,omitempty"`
	Integral int    `json:"integral,omitempty"`
	Desc     string `json:"pack_desc,omitempty"`
	Ctime    string `json:"ctime,omitempty"`
}

type UserBindInfo struct {
	MID    int64  `json:"mid"`
	Phone  int    `json:"phone"`
	Action string `json:"action"`
}

func (u *UnicomJson) UnicomJSONChange() (err error) {
	if u.Ordertypes != "" {
		if _, err = strconv.Atoi(u.Ordertypes); err != nil {
			log.Error("UnicomJsonChange error(%v)", u)
		}
	}
	return
}

func (u *UnicomIP) UnicomIPChange() {
	u.IPStartUint = u.unicomIPTOUint(u.Ipbegin)
	u.IPEndUint = u.unicomIPTOUint(u.Ipend)
}

// nolint:gomnd
func (u *UnicomIP) unicomIPTOUint(ip int) (ipUnit uint32) {
	var (
		ip1, ip2, ip3, ip4 int
		ipStr              string
	)
	var _initIP = "%d.%d.%d.%d"
	ip1 = ip / 1000000000
	ip2 = (ip / 1000000) % 1000
	ip3 = (ip / 1000) % 1000
	ip4 = ip % 1000
	ipStr = fmt.Sprintf(_initIP, ip1, ip2, ip3, ip4)
	ipUnit = model.InetAtoN(ipStr)
	return
}

func (t *UserLog) UserLogJSONChange(jsonData string) (err error) {
	if err = json.Unmarshal([]byte(jsonData), &t); err != nil {
		return
	}
	return
}

func (u *UserBindV2) UserBindDateChange(ctime, mtime time.Time) {
	u.BindingTime = ctime.Format("2006-01-02")
	u.UpdateTime = mtime.Format("2006-01-02")
}

type ActivateResult struct {
	Flag    string `json:"flag"`
	Product string `json:"product"`
	Package string `json:"package"`
}

type Activate struct {
	Usertype   int    `json:"usertype,omitempty"`
	Cardtype   int    `json:"cardtype,omitempty"`
	Flowtype   int    `json:"flowtype,omitempty"`
	Desc       string `json:"desc,omitempty"`
	ProductTag string `json:"product_tag,omitempty"`
}

type ActiveState struct {
	ProductID   string `json:"product_id"`
	TfType      int    `json:"tf_type"`
	TfWay       string `json:"tf_way"`
	ProductDesc string `json:"product_desc"`
	ProductTag  string `json:"product_tag"`
	ProductType int    `json:"product_type"`
	FakeID      string `json:"fake_id"`
	Usermob     string `json:"usermob"`
}

type UserActiveParam struct {
	Build        int64  `form:"build"`
	Platform     string `form:"platform"`
	Auto         bool   `form:"auto"`
	Usermob      string `form:"usermob"`
	Pip          string `form:"pip"`
	FakeID       string `form:"fake_id"`
	NeedFlowAuto bool   `form:"need_flow_auto"`
	Mid          int64  `form:"-"`
	Buvid        string `form:"-"`
	IP           string `form:"-"`
	SinglePip    string `form:"-"`
}

type UserActiveLog struct {
	Mid      int64
	Build    int64
	Buvid    string
	Platform string
	IP       string
	Type     string
	Result   string
	Suggest  string
}

type FakeIDInfoResponse struct {
	Result    string `json:"result"`
	SeqID     string `json:"seqid"`
	RespType  string `json:"resp_type"`
	RespCode  string `json:"resp_code"`
	OrderTime string `json:"order_time"` // 时间戳
	PCode     string `json:"pcode"`      // fake_id
}

type UserMobInfo struct {
	Usermob string `json:"usermob"`
	FakeID  string `json:"fake_id"`
	Period  int64  `json:"period"`
	Month   string `json:"month"`
}

type VerifyBiliBiliCardResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CouponParam struct {
	AssetRequest struct {
		Channel     int32  `json:"channel"`
		SourceBizId string `json:"sourceBizId"`
		Mid         int64  `json:"uid"`
	} `json:"assetRequest"`
	SourceAuthorityId string `json:"sourceAuthorityId"`
	SourceId          string `json:"sourceId"`
}

func (c CouponParam) Verify(v CouponParam) bool {
	return c == v
}

type UnicomFlowTryoutParam struct {
	FakeID string `form:"fake_id"`
	Pip    string `form:"pip"`
	IP     string `form:"-"`
}
