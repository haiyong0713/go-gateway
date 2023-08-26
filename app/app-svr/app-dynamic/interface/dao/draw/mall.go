package dao

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"
	mecode "go-gateway/ecode"

	"go-common/library/ecode"
	"go-common/library/log"

	"github.com/pkg/errors"
)

const (
	_urlMallSearch = "/mall/internal/search/searchEs"
)

func (d *Dao) SearchMallItems(ctx context.Context, mid uint64, word string, page, pageSize int) (items []*model.MallSearchItem, hasMore bool, err error) {
	page = fixPage(page)
	type term struct {
		Field  string   `json:"field"`
		Values []string `json:"values"`
	}
	var termQueries []*term
	termQueries = append(termQueries, &term{
		Values: []string{"1"},
		Field:  "verify_state"},
	)
	// 电商接口入餐是整个json string，
	// 无法利用bm的http client
	params := map[string]interface{}{}
	params["mid"] = strconv.Itoa(int(mid))
	params["page"] = strconv.Itoa(page)
	params["pagesize"] = strconv.Itoa(pageSize)
	params["sort_type"] = "totalrank"
	params["sort_order"] = "asc"
	params["termQueries"] = termQueries
	params["keyword"] = word
	paramsb, _ := json.Marshal(params)
	var ret struct {
		Code     int    `json:"code"`
		NumPages int    `json:"numPages"`
		Page     int    `json:"page"`
		Msg      string `json:"msg"`
		Result   struct {
			Product []*model.MallSearchItem
		} `json:"result"`
	}
	url := d.conf.Hosts.MallCo + _urlMallSearch
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(paramsb)))
	if err != nil {
		log.Error("%s create request failed, param:(%s)", _urlMainSearch, string(paramsb))
		return nil, hasMore, mecode.ParamInvalid
	}
	req = req.WithContext(ctx)
	if err = d.client.JSON(ctx, req, &ret); err != nil {
		log.Error("%s query failed, params:(%s) error(%v)", _urlMallSearch, string(paramsb), err)
		return nil, hasMore, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s return error(%d) msg(%s) params(%s)", _urlMallSearch, ret.Code, ret.Msg, string(paramsb))
		err = errors.Wrap(ecode.Int(ret.Code), url+"?"+string(paramsb))
		return nil, hasMore, err
	}
	items = ret.Result.Product
	if ret.Page < ret.NumPages {
		hasMore = true
	}
	return
}
