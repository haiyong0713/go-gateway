package model

import (
	"fmt"
	xtime "go-common/library/time"
	"math/big"
	"net"
)

type RecordState int

const (
	StateRecording RecordState = 0
	StateSucess    RecordState = 1
	StateCancel    RecordState = 2
)

type ISP string

const (
	ISPUnicom ISP = "cu"
	ISPTelcom ISP = "ct"
	ISPMobile ISP = "cm"
)

type RecordBusiness string

// Kratos hello kratos.
type Kratos struct {
	Hello string
}

type TFRecord struct {
	ISP        string `json:"isp,omitempty"`
	RemoteIP   string `json:"remote_ip,omitempty"`
	RemoteHost string `json:"remote_host,omitempty"`
	Count      int    `json:"count,omitempty"`
	Size       int    `json:"size,omitempty"`
	FullURI    string `json:"full_uri,omitempty"`
	Info       string `json:"info,omitempty"`
}

// Comment: 免流局数据备案表
type FreeRecord struct {
	// Comment: 主键ID
	ID int `json:"id"`
	// Comment: IP段起始
	IPStart string `json:"ip_start"`
	// Comment: IP段起始int型
	// Default: 0
	IPStartInt int64 `json:"ip_start_int"`
	// Comment: IP段结束
	IPEnd string `json:"ip_end"`
	// Comment: IP段结束int型
	// Default: 0
	IPEndInt int64 `json:"ip_end_int"`
	// Comment: 局数据备案所属运营商
	ISP ISP `json:"isp"`
	// Comment: 局数据是否是BGP，0-否,1-是
	// Default: 0
	IsBGP bool `json:"is_bgp"`
	// Comment: 业务
	Business RecordBusiness `json:"business"`
	// Comment: IP备案状态，0-备案中,1-备案成功,2-已下线
	// Default: 0
	State RecordState `json:"state"`
	// Comment: 备案成功时间
	// Default: 0000-00-00 00:00:00
	SuccessTime xtime.Time `json:"success_time"`
	// Comment: 备案下线时间
	// Default: 0000-00-00 00:00:00
	CancelTime xtime.Time `json:"cancel_time"`
	// Comment: 创建时间
	// Default: CURRENT_TIMESTAMP
	Ctime xtime.Time `json:"ctime"`
	// Comment: 最后修改时间
	// Default: CURRENT_TIMESTAMP
	Mtime      xtime.Time `json:"mtime"`
	CtimeHuman string     `json:"ctime_human"`
}

func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InetAtoN(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

func IsIPv4(ip string) bool {
	ipv := net.ParseIP(ip)
	if ip := ipv.To4(); ip != nil {
		return true
	} else {
		return false
	}
}

func CheckIP(ipStart, ipEnd string) bool {
	if !IsIPv4(ipStart) || !IsIPv4(ipEnd) {
		return false
	}
	if InetAtoN(ipStart) > InetAtoN(ipEnd) {
		return false
	}
	return true
}
