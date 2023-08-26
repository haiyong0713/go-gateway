package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/search"

	"github.com/pkg/errors"
)

const (
	_main    = "/main/search"
	_suggest = "/main/suggest/new"
)

type Dao struct {
	client   *httpx.Client
	main     string
	suggest  string
	redisCli *redis.Redis
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		client:   httpx.NewClient(c.HTTPSearch),
		main:     c.Host.Search + _main,
		suggest:  c.Host.Search + _suggest,
		redisCli: redis.NewRedis(c.Redis.Entrance),
	}
	return d
}

// Search 接口文档：https://git.bilibili.co/bili2_search/api_docs/-/blob/master/bsearch_new_api.md
func (d *Dao) Search(c context.Context, mid, rid int64, pn, ps int, keyword, buvid string) (*search.Search, error) {
	var (
		ip = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("from_car", "1")
	params.Set("keyword", keyword)
	params.Set("main_ver", "v3")
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("clientip", ip)
	params.Set("flow_need", "1")
	params.Set("search_type", "all")
	//请求平台 目前没有使用
	//params.Set("platform", "")
	if rid > 0 {
		params.Set("tids", strconv.FormatInt(rid, 10))
	}
	// 就第一页需要番剧信息后面都不要了
	if pn == 1 {
		params.Set("media_bangumi_num", "3")
		params.Set("media_ft_num", "3")
	}
	req, err := d.client.NewRequest("GET", d.main, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", buvid)
	var res *search.Search
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
	}
	return res, nil
}

// Suggest suggest data.
func (d *Dao) Suggest(c context.Context, plat int8, mid int64, platform, buvid, term, mobiApp, device string, build, highlight int) (*search.Suggest, error) {
	var (
		req *http.Request
		ip  = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("suggest_type", "accurate")
	params.Set("platform", platform)
	params.Set("mobi_app", "android")
	params.Set("clientip", ip)
	params.Set("highlight", strconv.Itoa(highlight))
	params.Set("build", strconv.Itoa(build))
	if mid != 0 {
		params.Set("userid", strconv.FormatInt(mid, 10))
	}
	params.Set("term", term)
	params.Set("sug_num", "10")
	params.Set("buvid", buvid)
	req, err := d.client.NewRequest("GET", d.suggest, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", buvid)
	res := &search.Suggest{}
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.suggest+"?"+params.Encode())
		return nil, err
	}
	return res, nil
}

func (d *Dao) Upper(c context.Context, mid int64, pn, ps int, keyword, buvid string) ([]*search.User, error) {
	var (
		ip = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("is_bvid", "1")
	params.Set("main_ver", "v3")
	params.Set("from_car", "1")
	params.Set("keyword", keyword)
	params.Set("userid", strconv.FormatInt(mid, 10))
	params.Set("search_type", "bili_user")
	params.Set("user_type", "0")
	params.Set("order", "totalrank")
	params.Set("order_sort", "0")
	params.Set("func", "search")
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("smerge", "1")
	req, err := d.client.NewRequest("GET", d.main, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", buvid)
	var res struct {
		Code   int            `json:"code"`
		SeID   string         `json:"seid"`
		Pages  int            `json:"numPages"`
		ExpStr string         `json:"exp_str"`
		List   []*search.User `json:"result"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.main+"?"+params.Encode())
	}
	return res.List, nil
}

func (d *Dao) GetRegionOffsetCacheById(ctx context.Context, id string) (res string, err error) {
	conn := d.redisCli.Conn(ctx)
	defer conn.Close()
	key := keyRegionBuvid(id)
	if res, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			log.Warnc(ctx, "GetRegionOffsetCacheById(%v) err: %v", key, err)
			err = nil
		} else {
			log.Errorc(ctx, "GetRegionOffsetCacheById(%v) err: %v", key, err)
		}
	}
	return
}

func (d *Dao) SaveRegionOffsetCache(ctx context.Context, id string, offset string) (ok bool, err error) {
	conn := d.redisCli.Conn(ctx)
	defer conn.Close()
	key := keyRegionBuvid(id)
	_, err = conn.Do("SET", key, offset, "EX", 60*60*12)
	if err != nil {
		log.Errorc(ctx, "SaveRegionOffsetCache conn.Do(SET(%s)) error(%v)", key, err)
		return
	}
	return
}

func keyRegionBuvid(id string) string {
	if id == "" {
		id = "null"
	}
	return fmt.Sprintf("xp::region::%s:", id)
}
