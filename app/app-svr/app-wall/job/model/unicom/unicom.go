package unicom

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-wall/job/model"
)

type FreeProduct string

var (
	FlowProduct = FreeProduct("flow")
	CardProduct = FreeProduct("card")
)

// ComicView def.
type ComicView struct {
	Uid     int64  `json:"uid"`
	ComicID int64  `json:"comic_id"`
	EpID    int64  `json:"ep_id"`
	Time    string `json:"time"`
}

type UserBind struct {
	Usermob  string    `json:"usermob,omitempty"`
	Phone    int       `json:"phone"`
	Mid      int64     `json:"mid"`
	State    int       `json:"state,omitempty"`
	Integral int       `json:"integral"`
	Flow     int       `json:"flow"`
	Monthly  time.Time `json:"monthly"`
}

// type

type ClickMsg struct {
	Plat       int8
	AID        int64
	MID        int64
	Lv         int8
	BvID       string
	CTime      int64
	STime      int64
	IP         string
	KafkaBs    []byte
	EpID       int64
	SeasonType int
	UserAgent  string
}

type Unicom struct {
	ID          int        `json:"-"`
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

type UnicomIP struct {
	Ipbegin     int    `json:"-"`
	Ipend       int    `json:"-"`
	IPStartUint uint32 `json:"-"`
	IPEndUint   uint32 `json:"-"`
}

type UserPackLog struct {
	Phone     int    `json:"-"`
	Usermob   string `json:"-"`
	Mid       int64  `json:"-"`
	RequestNo string `json:"-"`
	Type      int    `json:"-"`
	Desc      string `json:"-"`
	Integral  int    `json:"-"`
	UserDesc  string `json:"-"`
}

func (u *UnicomIP) UnicomIPStrToint(ipstart, ipend string) {
	u.Ipbegin = ipToInt(ipstart)
	u.Ipend = ipToInt(ipend)
}

// ipToint
// nolint:gomnd
func ipToInt(ipString string) (ipInt int) {
	tmp := strings.Split(ipString, ".")
	if len(tmp) < 4 {
		return
	}
	var ipStr string
	for _, tip := range tmp {
		var (
			ipLen = len(tip)
			last  int
			ip1   string
		)
		if ipLen < 3 {
			last = 3 - ipLen
			switch last {
			case 1:
				ip1 = "0" + tip
			case 2:
				ip1 = "00" + tip
			case 3:
				ip1 = "000"
			}
		} else {
			ip1 = tip
		}
		ipStr = ipStr + ip1
	}
	ipInt, _ = strconv.Atoi(ipStr)
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
