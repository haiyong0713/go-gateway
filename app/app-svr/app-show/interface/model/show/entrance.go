package show

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/app-show/interface/model"
)

const (
	_oneTimeRedDot = 1 // 一次性初始值红点
	_regularRedDot = 2 // 接口更新周期性红点
)

// EntranceCore .
type EntranceCore struct {
	Icon        string
	Grey        int
	Title       string
	Version     int32
	UpdateTime  time.Time
	ModuleID    string
	RedirectURI string
	ID          int64
	TopPhoto    string
	ShareInfo
}

// EntranceDB db 中的原始数据结构 .
type EntranceDB struct {
	EntranceCore
	RedDotText string
	WhiteList  string
	BlackList  string
	RedDot     int
	BuildLimit string
	BGroup     BGroup
}

// 人群包
type BGroup struct {
	Business string
	Name     string
}

// EntranceMem 内存中的入口数据
type EntranceMem struct {
	EntranceCore
	VerCtrls         map[int8]*EntranceVerCtrl // 版本过滤信息
	PopularWhitelist map[int64]struct{}
	PopularBlacklist map[int64]struct{}
	Show             *EntranceShow // 展示出去的入口数据
	BGroup           BGroup        //人群包信息
}

// EntranceVerCtrl get from db .
type EntranceVerCtrl struct {
	Plat           int8   `json:"plat"`
	ConditionStart string `json:"condition_start"`
	BuildStart     int    `json:"build_start"`
	ConditionEnd   string `json:"condition_end"`
	BuildEnd       int    `json:"build_end"`
}

// FromEntranceDB .
func (v *EntranceMem) FromEntranceDB(a *EntranceDB) (err error) {
	v.EntranceCore = a.EntranceCore
	v.BGroup = a.BGroup
	if a.BuildLimit != "" { // 版本控制
		var res []*EntranceVerCtrl
		if err = json.Unmarshal([]byte(a.BuildLimit), &res); err != nil {
			log.Error("[FromSource] json.Unmarshal() error(%v)", err)
			return
		}
		v.VerCtrls = make(map[int8]*EntranceVerCtrl, len(res))
		for _, ctrl := range res {
			v.VerCtrls[ctrl.Plat] = ctrl
		}
	}
	v.PopularWhitelist = dealMidList(a.WhiteList) // 黑白名单
	v.PopularBlacklist = dealMidList(a.BlackList)
	v.Show = &EntranceShow{
		ModuleID:   a.ModuleID,
		Icon:       a.Icon,
		Title:      a.Title,
		URI:        a.RedirectURI,
		EntranceID: a.ID,
		TopPhoto:   a.TopPhoto,
		ShareInfo: ShareInfo{
			CurrentTopPhoto: a.TopPhoto,
			ShareDesc:       a.ShareDesc,
			ShareTitle:      a.ShareTitle,
			ShareSubTitle:   a.ShareSubTitle,
			ShareIcon:       a.ShareIcon,
		},
	}
	if a.RedDot == _oneTimeRedDot || (a.RedDot == _regularRedDot && a.Version != 0) { // 一次性红点 或者 接口红点version不为0
		v.Show.Bubble = &Bubble{
			BubbleContent: a.RedDotText,
			Version:       a.Version,
			Stime:         a.UpdateTime.Time().Unix(),
		}
	}
	return
}

// dealMidList deal black and white list .
func dealMidList(midStr string) (res map[int64]struct{}) {
	if midStr == "" {
		return
	}
	res = make(map[int64]struct{})
	mids := strings.Split(midStr, ",")
	for _, v := range mids {
		if mid, _ := strconv.ParseInt(v, 10, 64); mid != 0 {
			res[mid] = struct{}{}
		}
	}
	return
}

type TopEntranceFilterMeta struct {
	Mid          int64
	Buvid        string
	Build        int
	MobiApp      string
	Device       string
	BGroupResult map[string]bool
}

// CanShow 处理版本过滤、黑名单、白名单、灰度逻辑
func (v *EntranceMem) CanShow(in TopEntranceFilterMeta) bool {
	plat := model.Plat2(in.MobiApp, in.Device)
	ctrl, ok := v.VerCtrls[plat]
	if !ok { // 没有相应版本配置，则不出
		return false
	}
	// 未通过版本过滤
	if !(ctrl.PassCtrl(in.Build) || plat == model.PlatH5) {
		return false
	}
	if _, ok := v.PopularBlacklist[in.Mid]; ok {
		return false
	}
	if _, ok := v.PopularWhitelist[in.Mid]; ok {
		return true
	}
	if v.BGroup.Business != "" && v.BGroup.Name != "" && in.BGroupResult[BGroupKey(v.BGroup.Business, v.BGroup.Name)] {
		return true
	}
	return v.Grey == 100 || crc32.ChecksumIEEE([]byte(in.Buvid))%100 < uint32(v.Grey)
}

func BGroupKey(business string, name string) string {
	return fmt.Sprintf("%s_%s", business, name)
}

// PassCtrl 处理版本过滤逻辑
func (v *EntranceVerCtrl) PassCtrl(build int) bool {
	if v.ConditionStart != "" && v.BuildStart != 0 {
		switch v.ConditionStart {
		case "gt":
			if build <= v.BuildStart {
				return false
			}
		case "ge":
			if build < v.BuildStart {
				return false
			}
		}
	}
	if v.ConditionEnd != "" && v.BuildEnd != 0 {
		switch v.ConditionEnd {
		case "lt":
			if build >= v.BuildEnd {
				return false
			}
		case "le":
			if build > v.BuildEnd {
				return false
			}
		}
	}
	return true
}
