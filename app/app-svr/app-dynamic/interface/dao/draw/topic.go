package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"

	"github.com/pkg/errors"
)

const (
	_urlGetHotTopic = "/topic/v1/Topic/hots"
)

func (d *Dao) GetHotTopicTopK(ctx context.Context, k int) (topics []*model.TopicHotItem, err error) {
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			List []*model.TopicHotItem `json:"list"`
		} `json:"data"`
	}
	queryUrl := d.conf.Hosts.VcCo + _urlGetHotTopic
	if err = d.client.Get(ctx, queryUrl, "", nil, &ret); err != nil {
		log.Error("%s query failed, error(%v)", _urlGetHotTopic, err)
		return nil, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s return error(%d) msg(%s)", _urlGetHotTopic, ret.Code, ret.Msg)
		err = errors.Wrap(ecode.Int(ret.Code), queryUrl)
		return nil, err
	}
	return ret.Data.List[:model.MinInt(len(ret.Data.List), k)], nil
}

func (d *Dao) SearchTopic(ctx context.Context, word string, page, pageSize int) (topics []*model.TopicSearchItem, hasMore bool, err error) {
	page = fixPage(page)
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("search_type", "tag")
	params.Set("keyword", word)
	params.Set("page", strconv.Itoa(page))
	params.Set("page_size", strconv.Itoa(pageSize))
	var ret struct {
		Code     int                      `json:"code"`
		Result   []*model.TopicSearchItem `json:"result"`
		NumPages int                      `json:"numPages"`
		Page     int                      `json:"page"`
	}
	queryUrl := d.conf.Hosts.SearchCo + _urlMainSearch
	if err = d.client.Get(ctx, queryUrl, "", params, &ret); err != nil {
		log.Error("%s tag query failed, params:(%s) error(%v)", _urlMainSearch, params.Encode(), err)
		return nil, hasMore, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s tag return err code(%d) params:(%s)", _urlMainSearch, ret.Code, params.Encode())
		err = errors.Wrap(ecode.Int(ret.Code), queryUrl+"?"+params.Encode())
		return nil, hasMore, err
	}
	if ret.Page < ret.NumPages {
		hasMore = true
	}
	if len(ret.Result) == 0 {
		log.Error("%s tag search return empty, params:(%s)", _urlMainSearch, params.Encode())
		return
	}
	topics = ret.Result
	return
}
