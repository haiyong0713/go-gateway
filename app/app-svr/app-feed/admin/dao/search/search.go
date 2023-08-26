package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/dataplat"
	searchModel "go-gateway/app/app-svr/app-feed/admin/model/search"
)

// SetSearchAuditStat set hot publish state to MC
func (d *Dao) SetSearchAuditStat(c context.Context, key string, state bool) (err error) {
	//nolint:gosimple
	var (
		p searchModel.PublishState
	)
	p = searchModel.PublishState{
		Date:  time.Now().Format("2006-01-02"),
		State: state,
	}
	itemJSON := &memcache.Item{
		Key:        key,
		Flags:      memcache.FlagJSON,
		Object:     p,
		Expiration: 0,
	}
	return d.MC.Set(c, itemJSON)
}

// GetSearchAuditStat get hot publish state from MC
func (d *Dao) GetSearchAuditStat(c context.Context, key string) (f bool, date string, err error) {
	var (
		p searchModel.PublishState
	)
	if err = d.MC.Get(c, key).Scan(&p); err == memcache.ErrNotFound {
		return false, "", nil
	}
	return p.State, p.Date, nil
}

// query在最终url中是json格式
// 20210714 clickhouse迁移，搜索热词修改clusterName
func (d *Dao) CallDataAPI(c context.Context, api string, query *dataplat.Query, res interface{}) (err error) {
	var response = &dataplat.Response{
		Result: res,
	}
	if query.Error() != nil {
		err = query.Error()
		log.Error("query error, err=%s", err)
		return
	}
	var params = url.Values{}
	params.Add("query", query.String())

	if err = d.DataPlatClient2.Get(c, api, params, response); err != nil {
		log.Error("fail to get response, err=%+v", err)
		return
	}

	if response.Code != http.StatusOK {
		err = fmt.Errorf("code:%d, msg:%s", response.Code, response.Msg)
		return
	}
	return
}

// query在最终url中是普通sql格式
// 20210714 clickhouse迁移，搜索热词修改clusterName
func (d *Dao) CallDataAPI_normal(c context.Context, api string, query string, res interface{}) (err error) {
	var response = &dataplat.Response2{
		Results: res,
	}

	var params = url.Values{}
	params.Add("query", query)
	if err = d.DataPlatClient2.Get(c, api, params, response); err != nil {
		log.Error("fail to get response, err=%+v", err)
		return
	}

	if response.Code != http.StatusOK {
		err = fmt.Errorf("code:%d, msg:%s", response.Code, response.Msg)
		return
	}
	return
}
