package common

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

const (
	logURL = "/x/admin/search/log"
)

// LogAction log action
func (s *Service) LogAction(c context.Context, param *common.Log) (res *common.LogManagers, err error) {
	var (
		logS  []*common.LogSearch
		items []*common.LogManager
	)
	res = &common.LogManagers{}
	params := url.Values{}
	params.Set("appid", "log_audit")
	params.Set("business", strconv.FormatUint(common.BusinessID, 10))
	params.Set("order", "ctime")
	params.Set("type", strconv.FormatInt(param.Type, 10))
	params.Set("ps", strconv.FormatInt(param.Ps, 10))
	params.Set("pn", strconv.FormatInt(param.Pn, 10))
	params.Set("ctime_from", param.Starttime)
	params.Set("ctime_to", param.Endtime)
	params.Set("uname", param.Uname)
	params.Set("action", param.Action)
	params.Set("str_1", param.Title)
	if param.ID != 0 {
		params.Set("oid", strconv.FormatInt(param.ID, 10))
	}
	if param.Query != "" {
		params.Set("str_0", param.Query)
	}
	l := &common.LogES{
		Data: &common.SearchResult{
			Page: &common.SearchPage{},
		},
	}
	if err = s.client.Get(c, s.managerURL+logURL, "", params, l); err != nil {
		return
	}
	for _, v := range l.Data.Result {
		search := &common.LogSearch{}
		if err = json.Unmarshal(v, search); err != nil {
			log.Error("LogAction.json.Unmarshal(%s)  error(%v)", string(v), err)
			continue
		}
		search.ActionEn = search.Action
		search.Action = s.ActionName(search.Action)
		logS = append(logS, search)
	}
	for _, v := range logS {
		tmp := &common.LogManager{
			ID:        v.OID,
			OID:       v.OID,
			Uname:     v.Uname,
			UID:       v.UID,
			Type:      v.Type,
			ExtraData: v.ExtraData,
			Action:    v.Action,
			CTime:     v.CTime,
			ActionEn:  v.ActionEn,
			Str_0:     v.Str_0,
			Str_1:     v.Str_1,
		}
		items = append(items, tmp)
	}
	res.Item = items
	res.Page.TotalItems = int(l.Data.Page.Total)
	res.Page.PageSize = l.Data.Page.Ps
	res.Page.CurrentPage = l.Data.Page.Pn
	return
}

func (s *Service) ActionName(action string) string {
	action = strings.ToLower(action)
	if strings.Contains(action, "add") {
		return "添加"
	} else if strings.Contains(action, "del") {
		return "删除"
	} else if strings.Contains(action, "up") {
		return "更新"
	} else if strings.Contains(action, "pub") {
		return "发布"
	} else if strings.Contains(action, "opt") {
		return "审核"
	} else if strings.Contains(action, "refuse") {
		return "拒绝"
	} else if strings.Contains(action, "online") {
		return "上线"
	} else if strings.Contains(action, "offline") || strings.Contains(action, "hidden") {
		return "下线"
	} else if strings.Contains(action, "pass") {
		return "通过"
	} else if strings.Contains(action, "re_audit") {
		return "重新审核"
	} else if strings.Contains(action, "forbidden") {
		return "禁用"
	} else if strings.Contains(action, "activation") {
		return "启用"
	}
	return ""
}
