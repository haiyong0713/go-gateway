package cards

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/log"
	cards "go-gateway/app/web-svr/activity/interface/model/cards"
)

// FollowMid 获取关注mid
func (d *Dao) FollowMid(c context.Context, uri string, timeStamp int64) (res []*cards.FollowMid, err error) {
	params := url.Values{}
	params.Set("t", strconv.FormatInt(timeStamp, 10))
	res = make([]*cards.FollowMid, 0)
	if err = d.singleClient.Get(c, uri, "", params, &res); err != nil {
		log.Errorc(c, "FollowMid d.client.Get(%s) error(%+v)", uri+"?"+params.Encode(), err)
		return
	}
	log.Infoc(c, "FollowMid d.client.Get url(%s) params (%v) res(%v)", uri, params.Encode(), res)
	return
}

// OgvLink 获取关注mid
func (d *Dao) OgvLink(c context.Context, uri string, timeStamp int64) (res []*cards.OgvLink, err error) {
	params := url.Values{}
	params.Set("t", strconv.FormatInt(timeStamp, 10))
	res = make([]*cards.OgvLink, 0)
	if err = d.singleClient.Get(c, uri, "", params, &res); err != nil {
		log.Errorc(c, "OgvLink d.client.Get(%s) error(%+v)", uri+"?"+params.Encode(), err)
		return
	}
	log.Infoc(c, "OgvLink d.client.Get url(%s) params (%v) res(%v)", uri, params.Encode(), res)
	return
}
