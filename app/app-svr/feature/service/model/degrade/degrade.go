package degrade

import (
	"encoding/json"
	"strings"

	"go-common/library/log"
)

const (
	RuleNone    = -1
	DefLogLevel = 4
)

type DisplayLimitRes struct {
	DisplayType string        `json:"display_type"`
	LimitType   string        `json:"limit_type"`
	LimitList   []string      `json:"limit_list"`
	Rules       *DegradeRange `json:"rules"`
	Rank        int64         `json:"rank"`
	Direction   int           `json:"direction"` // 黑白名单	1：白	2：黑
}

type DegradeRange struct {
	SysGte   int64   `json:"sys_gte"`
	SysLte   int64   `json:"sys_lte"`
	BuildGte int64   `json:"build_gte"`
	BuildLte int64   `json:"build_lte"`
	StoreGte int64   `json:"store_gte"`
	StoreLte int64   `json:"store_lte"`
	Enlarge  float32 `json:"enlarge"`
	LogLevel int32   `json:"log_level"`
}

type Range struct {
	SysVerRange [][]int64
	BuildRange  [][]int64
	LogLevel    int32
	Direction   int `json:"direction"` // 黑白名单	1：白	2：黑
}

// 该值是否在区间内
func InIntervals(itvs [][]int64, val int64) bool {
	if len(itvs) == 0 { // 若为空，认为全局有效
		return true
	}
	for _, arr := range itvs {
		if val >= arr[0] && (val <= arr[1] || arr[1] == RuleNone) {
			return true
		}
	}
	return false
}

// map[渠道/品牌/机型] map[具体的渠道/品牌/机型名] *model.Range
func ToMap(displayLimits []*DisplayLimitRes) map[string]map[string]*Range {
	res := make(map[string]map[string]*Range)
	for _, l := range displayLimits {
		if res[l.LimitType] == nil {
			res[l.LimitType] = make(map[string]*Range)
		}
		for _, limitItem := range l.LimitList {
			if res[l.LimitType][limitItem] == nil {
				res[l.LimitType][limitItem] = new(Range)
			}
			res[l.LimitType][limitItem].SysVerRange = insert(res[l.LimitType][limitItem].SysVerRange, []int64{l.Rules.SysGte, l.Rules.SysLte})
			res[l.LimitType][limitItem].BuildRange = insert(res[l.LimitType][limitItem].BuildRange, []int64{l.Rules.BuildGte, l.Rules.BuildLte})
			res[l.LimitType][limitItem].LogLevel = l.Rules.LogLevel
			res[l.LimitType][limitItem].Direction = l.Direction
		}
	}
	return res
}

// 将新区间插入原有区间，返回一个新的有序的范围
// nolint:gomnd
func insert(itvs [][]int64, new []int64) [][]int64 {
	// 非法直接返回
	if len(new) != 2 {
		return itvs
	}
	// 之前是空的，直接加入
	if len(itvs) < 1 {
		return [][]int64{new}
	}
	// 如果之前是[-1,-1],直接返回
	if len(itvs) == 1 && itvs[0][0] == RuleNone && itvs[0][1] == RuleNone {
		return itvs
	}
	// 如果新插入的是[-1,-1]，直接覆盖
	if new[0] == RuleNone && new[1] == RuleNone {
		return [][]int64{new}
	}
	var (
		res = make([][]int64, 0)
		i   = 0
		n   = len(itvs)
	)
	for i < n && itvs[i][0] < new[0] {
		res = append(res, itvs[i])
		i++
	}
	if len(res) == 0 || res[len(res)-1][1] < new[0] {
		res = append(res, new)
	} else {
		if new[1] > res[len(res)-1][1] {
			res[len(res)-1][1] = new[1]
		}
	}
	for i < n {
		s := itvs[i][0]
		e := itvs[i][1]
		lastE := res[len(res)-1][1]
		if lastE < s {
			res = append(res, itvs[i])
		} else {
			if e > res[len(res)-1][1] {
				res[len(res)-1][1] = e
			}
		}
		i++
	}
	return res
}

type DisplayLimitDB struct {
	Id          int64  `json:"id"`
	DisplayType string `json:"display_type"`
	LimitType   string `json:"limit_type"` // chid/model/brand
	LimitList   string `json:"limit_list"`
	Rules       string `json:"rules"`
	Rank        int64  `json:"rank"`
	Direction   int    `json:"direction"` // 黑白名单	1：白	2：黑
}

func (v *DisplayLimitDB) ToRes() *DisplayLimitRes {
	rule := new(DegradeRange)
	if err := json.Unmarshal([]byte(v.Rules), &rule); err != nil {
		log.Warn("degradeWarn ToRes() req(%+v) err(%+v)", v, err)
		return nil
	}
	return &DisplayLimitRes{
		DisplayType: v.DisplayType,
		LimitType:   v.LimitType,
		LimitList:   strings.Split(v.LimitList, ","),
		Rules:       rule,
		Rank:        v.Rank,
		Direction:   v.Direction,
	}
}
