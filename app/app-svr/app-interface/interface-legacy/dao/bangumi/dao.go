package bangumi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/bangumi"

	pgcmedia "git.bilibili.co/bapis/bapis-go/pgc/servant/media"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcappcard "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	pgcfollow "git.bilibili.co/bapis/bapis-go/pgc/service/follow/media"
	pgcreview "git.bilibili.co/bapis/bapis-go/pgc/service/review"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	pgcstat "git.bilibili.co/bapis/bapis-go/pgc/service/stat/v1"

	"github.com/pkg/errors"
)

const (
	_season     = "/api/inner/season"
	_movie      = "/internal_api/movie_aid_info"
	_bp         = "/sponsor/inner/xAjaxGetBP"
	_card       = "/pgc/internal/season/search/card"
	_favDisplay = "/pgc/internal/follow/app/tab/view"
)

// Dao is bangumi dao
type Dao struct {
	client     *httpx.Client
	season     string
	movie      string
	bp         string
	card       string
	favDisplay string
	// grpc
	rpcClient           seasongrpc.SeasonClient
	pgcsearchClient     pgcsearch.SearchClient
	pgcinlineClient     pgcinline.InlineCardClient
	pgcstatClient       pgcstat.StatServiceClient
	pgcepClient         episodegrpc.EpisodeClient
	pgcCardClient       pgccardgrpc.CardClient
	pgcAppCardClient    pgcappcard.AppCardClient
	pgcMediaClient      pgcmedia.MediaClient
	pgcReviewClient     pgcreview.ReviewClient
	pgcReviewUserClient pgcreview.ReviewUserClient
	pgcFollowClient     pgcfollow.MediaFollowClient
}

// New bangumi dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:     httpx.NewClient(c.HTTPBangumi),
		season:     c.Host.Bangumi + _season,
		movie:      c.Host.Bangumi + _movie,
		bp:         c.Host.Bangumi + _bp,
		card:       c.Host.APICo + _card,
		favDisplay: c.Host.APICo + _favDisplay,
	}
	var err error
	if d.rpcClient, err = seasongrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("seasongrpc NewClientt error (%+v)", err))
	}
	if d.pgcsearchClient, d.pgcinlineClient, err = newClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcsearch pgcinline newClient error (%+v)", err))
	}
	if d.pgcstatClient, err = newStatClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcstat newStatClient error (%+v)", err))
	}
	if d.pgcepClient, err = episodegrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcep NewClient error (%+v)", err))
	}
	if d.pgcCardClient, err = pgccardgrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgccard NewClient error (%+v)", err))
	}
	if d.pgcAppCardClient, err = pgcappcard.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcAppCardClient NewClient error (%+v)", err))
	}
	if d.pgcMediaClient, err = pgcmedia.NewClientMedia(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcMediaClient NewClient error (%+v)", err))
	}
	if d.pgcReviewClient, err = pgcreview.NewClientReview(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcReviewClient NewClient error (%+v)", err))
	}
	if d.pgcReviewUserClient, err = pgcreview.NewClientReviewUser(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcReviewUserClient NewClient error (%+v)", err))
	}
	if d.pgcFollowClient, err = pgcfollow.NewClientMediaFollow(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("pgcFollowClient NewClient error (%+v)", err))
	}
	return
}

// Season bangumi Season.
func (d *Dao) Season(c context.Context, aid, mid int64, ip string) (s *bangumi.Season, err error) {
	params := url.Values{}
	params.Set("aid", strconv.FormatInt(aid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("type", "av")
	params.Set("build", "app-interface")
	params.Set("platform", "Golang")
	var res struct {
		Code   int             `json:"code"`
		Result *bangumi.Season `json:"result"`
	}
	if err = d.client.Get(c, d.season, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.season+"?"+params.Encode())
		return
	}
	s = res.Result
	return
}

// BPInfo get bp info data.
func (d *Dao) BPInfo(c context.Context, aid, mid int64, ip string) (data json.RawMessage, err error) {
	params := url.Values{}
	params.Set("aid", strconv.FormatInt(aid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("build", "app-interface")
	params.Set("platform", "Golang")
	var res struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err = d.client.Get(c, d.bp, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.bp+"?"+params.Encode())
		return
	}
	data = res.Data
	return
}

// Movie bangumi Movie
func (d *Dao) Movie(c context.Context, aid, mid int64, build int, mobiApp, device, ip string) (m *bangumi.Movie, err error) {
	params := url.Values{}
	params.Set("id", strconv.FormatInt(aid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("device", device)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", "Golang")
	var res struct {
		Code   int            `json:"code"`
		Result *bangumi.Movie `json:"result"`
	}
	if err = d.client.Get(c, d.movie, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.movie+"?"+params.Encode())
		return
	}
	m = res.Result
	return
}

// Concern ogv追番接口，由grpc接口MyFollows提供
func (d *Dao) Concern(ctx context.Context, mid, vmid int64, pn, ps int) (*pgcappcard.FollowReply, error) {
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

// Card bangumi card.
func (d *Dao) Card(c context.Context, mid int64, sids []int64) (s map[string]*bangumi.Card, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("season_ids", xstr.JoinInts(sids))
	var res struct {
		Code   int                      `json:"code"`
		Result map[string]*bangumi.Card `json:"result"`
	}
	if err = d.client.Get(c, d.card, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.card+"?"+params.Encode())
		return
	}
	s = res.Result
	return
}

// FavDisplay fav tab display or not.
func (d *Dao) FavDisplay(c context.Context, mid int64) (bangumi, cinema int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code   int `json:"code"`
		Result struct {
			Bangumi int `json:"bangumi"`
			Cinema  int `json:"cinema"`
		} `json:"result"`
	}
	if err = d.client.Get(c, d.favDisplay, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.favDisplay+"?"+params.Encode())
		return
	}
	bangumi = res.Result.Bangumi
	cinema = res.Result.Cinema
	return
}
