package model

import (
	"fmt"
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
)

var conditionMap = map[string]struct{}{
	_all:          {},
	_greaterEqual: {},
	_greaterThan:  {},
	_lessThan:     {},
	_lessEqual:    {},
}

var platformMap = map[string]struct{}{
	_platAndroid: {},
	_platIos:     {},
	_platOtt:     {},
}

type UploadReply struct {
	URL string `json:"url"`
}

// ChronosRule 后台请求的规则
type ChronosRule struct {
	Title      string      `json:"title"`
	Avids      string      `json:"avids"`
	Mids       string      `json:"mids"`
	BuildLimit []*VerLimit `json:"build_limit"`
	Gray       int32       `json:"gray"`
	File       string      `json:"file"`
}

// PlayerRule 是给app-player准备的结构
type PlayerRule struct {
	ChronosRule
	MD5 string `json:"md5"`
}

// 校验版本限制之间没有冲突
func (v *ChronosRule) BuildValidate() bool {
	var builds = make(map[string]*Build) // 按照platform区分build

	if len(v.BuildLimit) == 0 {
		v.BuildLimit = make([]*VerLimit, 0)
		return true
	}

	for _, v := range v.BuildLimit {
		if _, ok := conditionMap[v.Condition]; !ok {
			return false
		}
		if _, ok := platformMap[v.Platform]; !ok {
			return false
		}
		if v.Condition == _all { // all不检查冲突
			continue
		}
		if _, ok := builds[v.Platform]; !ok {
			builds[v.Platform] = new(Build)
		}
		val := int(v.Value)
		switch v.Condition { // 复用原有的检查代码
		case _lessEqual:
			builds[v.Platform].LE = val
		case _lessThan:
			builds[v.Platform].LT = val
		case _greaterEqual:
			builds[v.Platform].GE = val
		case _greaterThan:
			builds[v.Platform].GT = val
		}
	}
	for _, v := range builds {
		if !v.CheckRange() { // 未通过冲突检查
			return false
		}
	}
	return true
}

func idsValidate(ids string) bool {
	if ids == "" {
		return false
	}
	if ids == _all { // 全部可用
		return true
	}
	vals := strings.Split(ids, ",")
	if len(vals) == 0 { // ids不可为空
		return false
	}
	for _, v := range vals { // 逐个校验为数字且大于0
		if val, err := strconv.ParseInt(v, 10, 64); err != nil || val == 0 {
			return false
		}
	}
	return true
}

func RulesValidate(rules []*ChronosRule) (errorMsg string) {
	if len(rules) == 0 {
		return ""
	}
	names := make(map[string]struct{})
	for _, v := range rules {
		if _, ok := names[v.Title]; ok {
			return fmt.Sprintf("%s 标题出现重复", v.Title)
		}
		names[v.Title] = struct{}{}
		if !idsValidate(v.Avids) {
			return fmt.Sprintf("AVIDS %s 不合法", v.Avids)
		}
		if !idsValidate(v.Mids) {
			return fmt.Sprintf("MIDS %s 不合法", v.Mids)
		}
		if !v.BuildValidate() {
			return "版本限制不合法"
		}
		if v.Gray < 0 || v.Gray > 10000 {
			return "灰度值在0到10000之间"
		}
		if v.Avids == _all && v.Mids == _all && v.Gray == 10000 { // 危险rule
			if open := func() bool {
				for _, v := range v.BuildLimit {
					if v.Condition != _all {
						return false
					}
				}
				return true
			}(); open {
				log.Error("Chronos dangerous Rules %s", v.Title)
			}
		}
	}
	return ""
}

type VerLimit struct {
	Condition string `json:"condition"`
	Value     int64  `json:"value"`
	Platform  string `json:"platform"`
}
