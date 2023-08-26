package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"

	"github.com/pkg/errors"
)

const (
	_hotRcmdURI   = "/recommand"
	_hotRcmdCmd   = "hot"
	_hotTenprefix = "%d_hchashmap_car"

	_webPlat = 30
)

func (d *Dao) HotAiRcmd(c context.Context, mid int64, buvid string, pageNo, count int) (data []*model.HotItem, userFeature string, resCode int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("cmd", _hotRcmdCmd)
	params.Set("from", "10")
	timeout := time.Duration(d.c.HTTPClient.Read.Timeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("request_cnt", strconv.Itoa(count))
	params.Set("page_no", strconv.Itoa(pageNo))
	params.Set("web_hot", "1") // web热门
	params.Set("plat", strconv.FormatInt(_webPlat, 10))
	var res struct {
		Code        int              `json:"code"`
		Data        []*model.HotItem `json:"data"`
		UserFeature string           `json:"user_feature"`
	}
	if err = d.httpR.Get(c, d.hotRcmdURL, ip, params, &res); err != nil {
		return nil, "", ecode.ServerErr.Code(), err
	}
	if res.Code != ecode.OK.Code() {
		if res.Code == -3 { // code -3 热门数据已经拉到底了
			return []*model.HotItem{}, res.UserFeature, res.Code, nil
		}
		return nil, "", res.Code, errors.Wrap(ecode.Int(res.Code), d.hotRcmdURL+"?"+params.Encode())
	}
	return res.Data, res.UserFeature, res.Code, nil
}

func getHotKey(i int) string {
	return fmt.Sprintf(_hotTenprefix, i)
}

func (d *Dao) PopularCardTenCache(c context.Context, i, index, ps int) ([]*model.PopularCard, error) {
	var (
		key  = getHotKey(i)
		conn = d.redisPopular.Get(c)
		bss  [][]byte
	)
	defer conn.Close()
	arg := redis.Args{}.Add(key)
	for id := 0; id < ps; id++ {
		arg = arg.Add(index + id)
	}
	bss, err := redis.ByteSlices(conn.Do("HMGET", arg...))
	if err != nil {
		log.Error("PopularCardTenCache conn.Do(HGET,%s) error(%v)", key, err)
		return nil, err
	}
	var cards []*model.PopularCard
	for _, bs := range bss {
		if len(bs) == 0 {
			continue
		}
		card := new(model.PopularCard)
		if err = json.Unmarshal(bs, card); err != nil {
			log.Error("PopularCardTenCache json.Unmarshal(%s) error(%v)", string(bs), err)
			continue
		}
		cards = append(cards, card)
	}
	return cards, nil
}
