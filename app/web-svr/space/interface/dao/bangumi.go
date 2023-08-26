package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	pgcappcard "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
)

const (
	_build               = "0"
	_platform            = "web"
	_bangumiConcernURI   = "/api/concern_season"
	_bangumiUnConcernURI = "/api/unconcern_season"
)

// BangumiList get bangumi sub list by mid.
func (d *Dao) BangumiList(ctx context.Context, mid, vmid int64, pn, ps int) (*pgcappcard.FollowReply, error) {
	const (
		_pgcFollowFromSpace         = 2
		_pgcFollowFollowTypeBangumi = 1
	)
	args := &pgcappcard.FollowReq{
		Mid:        mid,
		Pn:         int32(pn),
		Ps:         int32(ps),
		Tid:        vmid,
		From:       _pgcFollowFromSpace,
		FollowType: _pgcFollowFollowTypeBangumi,
	}
	res, err := d.pgcAppCardClient.MyFollows(ctx, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// BangumiConcern bangumi concern.
func (d *Dao) BangumiConcern(c context.Context, mid, seasonID int64) (err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("season_id", strconv.FormatInt(seasonID, 10))
	params.Set("build", _build)
	params.Set("platform", _platform)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.httpW.Post(c, d.bangumiConcernURL, ip, params, &res); err != nil {
		log.Error("d.httpW.Post(%s,%d) error(%v)", d.bangumiConcernURL, mid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpW.Post(%s,%d) error(%v)", d.bangumiConcernURL, mid, err)
		err = ecode.Int(res.Code)
	}
	return
}

// BangumiUnConcern bangumi cancel sub.
func (d *Dao) BangumiUnConcern(c context.Context, mid, seasonID int64) (err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("season_id", strconv.FormatInt(seasonID, 10))
	params.Set("build", _build)
	params.Set("platform", _platform)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.httpW.Post(c, d.bangumiUnConcernURL, ip, params, &res); err != nil {
		log.Error("d.httpW.Post(%s,%d) error(%v)", d.bangumiUnConcernURL, mid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpW.Post(%s,%d) error(%v)", d.bangumiUnConcernURL, mid, err)
		err = ecode.Int(res.Code)
	}
	return
}
