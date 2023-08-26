package model

import (
	"encoding/json"
	"fmt"
	"net/url"

	"go-common/library/log"
)

const (
	// goto
	GotoPGC = "pgc"
	GotoAv  = "av"

	// 客户端根据source_type类型也请求不同业务方接口
	EntranceCommonSearch = "common_search"

	// season类型 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧，7:综艺
	SeasonTypeBangumi     = 1
	SeasonTypeMovie       = 2
	SeasonTypeDocumentary = 3
	SeasonTypeGc          = 4
	SeasonTypeTv          = 5
	SeasonTypeZi          = 7

	// season付费状态 0-免费可看  1-VIP免费可看 2-VIP付费可看 3-其他
	SeasonFree = 0
	SeasonVip  = 1

	// ep付费状态 2-免费 6-付费,大会员免费 7-付费抢先,大会员免费 8-全付费观看 9-全付费抢先  12-霹雳付费 13-仅大会员可看
	EpFree     = 2
	EpVipFree  = 6
	EpVipFree2 = 7
	EpOnlyVip  = 13
)

var (
	ParamHandler = func(main interface{}, entrance, keyword string) func(uri string) string {
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("ParamHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("ParamHandler url.ParseQuery error(%v)", err)
				return uri
			}
			// 特殊参数用于进入接口列表做插入逻辑使用
			if main != nil {
				b, _ := json.Marshal(main)
				params.Set("param", string(b))
			}
			params.Set("sourceType", entrance)
			if keyword != "" {
				params.Set("keyword", keyword)
			}
			u.RawQuery = params.Encode()
			return u.String()
		}
	}
)

// FillURI deal app schema.
func FillURI(id, cid int64, f func(uri string) string) (uri string) {
	uri = fmt.Sprintf("bilithings://player?aid=%d&cid=%d", id, cid)
	if f != nil {
		uri = f(uri)
	}
	return
}

func SearchPrune(id, childID int64, gt string) interface{} {
	return map[string]interface{}{
		"goto":     gt,
		"id":       id,
		"child_id": childID,
	}
}
