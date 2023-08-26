package dao

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"

	"github.com/pkg/errors"
)

const _wxHotURI = "hot-weixin-card.json"

const (
	_wxTeenagerRcmdURI = "/recommand"
	_wxTeenagerRcmdCmd = "teenager"
)

func (d *Dao) WxHot(c context.Context) (list []*model.WxArchiveCard, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var res struct {
		Code int `json:"code"`
		List []*struct {
			Goto       string `json:"goto"`
			ID         int64  `json:"id"`
			Desc       string `json:"desc"`
			CornerMark int    `json:"corner_mark"`
		} `json:"list"`
	}
	if err = d.httpBigData.Get(c, d.wxHotURL, ip, url.Values{}, &res); err != nil {
		log.Error("d.httpBigData.Get(%s) error(%v)", d.wxHotURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpBigData.Get(%s) error(%v)", d.wxHotURL, err)
		err = ecode.Int(res.Code)
		return
	}
	for _, v := range res.List {
		if v.Goto == "av" && v.ID > 0 {
			list = append(list, &model.WxArchiveCard{ID: v.ID, Desc: v.Desc, CornerMark: v.CornerMark})
		}
	}
	return
}

func (d *Dao) GetTeenageRcmdCards(c context.Context, mid int64, buvid string, freshType, ps int) (list []*model.WXTeenageRcmdItem, code int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	timeout := time.Duration(d.c.HTTPClient.Read.Timeout) / time.Millisecond
	params.Set("cmd", _wxTeenagerRcmdCmd)
	params.Set("from", "7")
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Add("buvid", buvid)
	params.Add("mid", strconv.FormatInt(mid, 10))
	params.Add("request_cnt", strconv.Itoa(ps))
	params.Add("fresh_type", strconv.Itoa(freshType))
	var res struct {
		Code        int                        `json:"code"`
		Data        []*model.WXTeenageRcmdItem `json:"data"`
		UserFeature json.RawMessage            `json:"user_feature"`
	}
	if err = d.httpR.Get(c, d.wxTeenageRcmdURL, ip, params, &res); err != nil {
		return nil, ecode.ServerErr.Code(), err
	}
	if res.Code != ecode.OK.Code() {
		if res.Code == -3 {
			// 稿件不足时，有多少返回多少
			return res.Data, res.Code, nil
		}
		return nil, res.Code, errors.Wrap(ecode.Int(res.Code), d.hotRcmdURL+"?"+params.Encode())
	}
	return res.Data, res.Code, nil
}
