package view

import (
	"strconv"
	"strings"

	"go-common/library/log"
)

const (
	_all          = "all"
	_lessThan     = "lt"
	_lessEqual    = "le"
	_greaterEqual = "ge"
	_greaterThan  = "gt"
	_platAndroid  = "android"
	_platIos      = "ios"
	_platOtt      = "ott"
	MaxGray       = 10000
)

// ChronosRule .
type ChronosRule struct {
	Title      string      `json:"title"`
	Avids      string      `json:"avids"`
	Mids       string      `json:"mids"`
	BuildLimit []*VerLimit `json:"build_limit"`
	Gray       int32       `json:"gray"`
	File       string      `json:"file"`
	MD5        string      `json:"md5"`
}

// VerLimit .
type VerLimit struct {
	Condition string `json:"condition"`
	Value     int64  `json:"value"`
	Platform  string `json:"platform"`
}

type PlatformLimit struct {
	Condition string `json:"condition"`
	Value     int64  `json:"value"`
}

// ChronosReply .
type ChronosReply struct {
	Title      string
	AllAvids   bool
	Avids      []int64
	AllMids    bool
	Mids       []int64
	BuildLimit map[string][]*PlatformLimit
	Gray       int32
	File       string
	MD5        string
}

func FormatPlayRule(in *ChronosRule) *ChronosReply {
	rly := &ChronosReply{Title: in.Title, File: in.File, MD5: in.MD5}
	if in.Mids == _all {
		rly.AllMids = true
	} else {
		rly.Mids = formatStringToInt(in.Mids)
		if len(rly.Mids) == 0 { //没有一个mid符合条件，此规则作废
			log.Error("FormatPlayRule  no mids in error(%v)", in)
			return nil
		}
	}
	if in.Avids == _all {
		rly.AllAvids = true
	} else {
		rly.Avids = formatStringToInt(in.Avids)
		if len(rly.Avids) == 0 { //没有一个mid符合条件，此规则作废
			log.Error("FormatPlayRule  no Avids in error(%v)", in)
			return nil
		}
	}
	if in.Gray <= 0 || in.Gray > MaxGray {
		log.Error("FormatPlayRule  no Gray in error(%d)", in.Gray)
		return nil
	}
	rly.Gray = in.Gray
	if len(in.BuildLimit) == 0 {
		log.Error("FormatPlayRule no BuildLimit error(%v)", in)
		return nil
	}
	rly.BuildLimit = make(map[string][]*PlatformLimit)
	for _, v := range in.BuildLimit {
		// condition 不合法
		if v.Condition != _lessEqual && v.Condition != _lessThan && v.Condition != _greaterEqual && v.Condition != _greaterThan && v.Condition != _all {
			log.Error("FormatPlayRule no Condition error(%s)", v.Condition)
			return nil
		}
		if v.Platform != _platAndroid && v.Platform != _platIos && v.Platform != _platOtt {
			log.Error("FormatPlayRule no platform error(%s)", v.Platform)
			return nil
		}
		rly.BuildLimit[v.Platform] = append(rly.BuildLimit[v.Platform], &PlatformLimit{Condition: v.Condition, Value: v.Value})
	}
	return rly
}

func formatStringToInt(str string) (ids []int64) {
	rly := strings.Split(str, ",")
	if len(rly) == 0 {
		return
	}
	for _, v := range rly {
		id, e := strconv.ParseInt(v, 10, 64)
		if e != nil {
			log.Error("formatStringToInt str%s error(%v)", str, e)
			continue
		}
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return
}

func IsInIDs(ids []int64, id int64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func InvalidBuild(srcBuild, cfgBuild int64, cfgCond string) bool {
	switch cfgCond {
	case _greaterThan:
		if cfgBuild >= srcBuild {
			return true
		}
	case _greaterEqual:
		if cfgBuild > srcBuild {
			return true
		}
	case _lessThan:
		if cfgBuild <= srcBuild {
			return true
		}
	case _lessEqual:
		if cfgBuild < srcBuild {
			return true
		}
	case _all: //所有的都通过
		return false
	default: //无法识别的，不通过
		return true
	}
	return false
}

type RuleMeta struct {
	AppKey        string
	ServiceKey    string
	Mid           int64
	Aid           int64
	RomVersion    string
	NetType       int64
	DeviceType    string
	EngineVersion string
	Buvid         string
	Build         int64
	MobiApp       string
	Device        string
}

type PackageInfoReply struct {
	Url  string
	MD5  string
	Sign string
}

type PackageInfo struct {
	Rank          int64  `json:"rank"`
	AppKey        string `json:"app_key"`
	ServiceKey    string `json:"service_key"`
	ResourceUrl   string `json:"resource_url"`
	Gray          int64  `json:"gray"`
	BlackList     string `json:"black_list"`
	WhiteList     string `json:"white_list"`
	VideoList     string `json:"video_list"`
	RomVersion    string `json:"rom_version"`
	NetType       string `json:"net_type"`
	DeviceType    string `json:"device_type"`
	EngineVersion string `json:"engine_version"`
	BuildLimitExp string `json:"build_limit_exp"`
	Sign          string `json:"sign"`
	Md5           string `json:"md5"`
}

type ChronosPkgReq struct {
	ServiceKey    string
	EngineVersion string
	Aid           int64
}
