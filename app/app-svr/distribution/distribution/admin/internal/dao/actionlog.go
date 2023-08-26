package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	ac "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/actionlog"

	"go-common/library/log"
)

func (d *dao) LogAction(ctx context.Context, param *ac.Log) (*ac.LogManagers, error) {
	var (
		uri    = fmt.Sprintf("%s%s", d.actionHost, "/x/admin/search/log/user_action")
		params = url.Values{}
	)

	if param.Mid != 0 {
		params.Set("mid", strconv.FormatInt(param.Mid, 10))
	}
	if param.UserName != "" {
		params.Set("str_0_like", param.UserName)
	}
	if param.Type != -1 {
		params.Set("type", strconv.FormatInt(param.Type, 10))
	}
	params.Set("ps", strconv.FormatInt(param.Ps, 10))
	params.Set("pn", strconv.FormatInt(param.Pn, 10))
	params.Set("ctime_from", param.CtimeFrom)
	params.Set("ctime_to", param.CtimeTo)
	params.Set("sort", param.Sort)
	params.Set("order", "ctime")
	params.Set("business", strconv.FormatInt(param.Business, 10))

	res := &struct {
		Code int `json:"code"`
		Data struct {
			Order  string            `json:"order"`
			Sort   string            `json:"sort"`
			Result []json.RawMessage `json:"result"`
			Debug  string            `json:"debug"`
			Page   struct {
				Pn    int   `json:"num"`
				Ps    int   `json:"size"`
				Total int64 `json:"total"`
			} `json:"page"`
		} `json:"data"`
	}{}

	if err := d.bmClient.Get(ctx, uri, "", params, res); err != nil {
		return nil, err
	}

	var (
		logSearch       []*ac.LogSearch
		logManagerItems []*ac.LogManagerItem
		logManagers     = &ac.LogManagers{}
	)

	for _, v := range res.Data.Result {
		search := &ac.LogSearch{}
		if err := json.Unmarshal(v, search); err != nil {
			log.Error("LogAction.json.Unmarshal(%s)  error(%v)", string(v), err)
			continue
		}
		logSearch = append(logSearch, search)
	}

	for _, v := range logSearch {
		logManagerItem := &ac.LogManagerItem{
			Mid:       v.Mid,
			Type:      v.Type,
			UserName:  v.UserName,
			ExtraData: v.ExtraData,
			CTime:     v.CTime,
			Business:  v.Business,
		}
		logManagerItems = append(logManagerItems, logManagerItem)
	}

	logManagers.Item = logManagerItems
	logManagers.Pager.TotalItems = int(res.Data.Page.Total)
	logManagers.Pager.PageSize = res.Data.Page.Ps
	logManagers.Pager.CurrentPage = res.Data.Page.Pn

	return logManagers, nil
}
