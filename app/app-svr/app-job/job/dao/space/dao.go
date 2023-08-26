package space

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-job/job/conf"
	"go-gateway/app/app-svr/app-job/job/model/space"

	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"

	"github.com/pkg/errors"
)

// Dao is favorite dao
type Dao struct {
	c          *conf.Config
	client     *httpx.Client
	clientAsyn *httpx.Client
	audioList  string
	// redis
	redis    *redis.Pool
	interRds *redis.Pool
	// up service grpc
	upClient  upgrpc.UpArchiveClient
	artClient artclient.ArticleGRPCClient
	// comic
	upComic string
	// contribute cache
	expireContribute int32
}

// New initial favorite dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		client:     httpx.NewClient(c.HTTPClient),
		clientAsyn: httpx.NewClient(c.HTTPClientAsyn),
		audioList:  c.Host.APICo + _audioList,
		// redis
		redis:    redis.NewPool(c.Redis.Contribute.Config),
		interRds: redis.NewPool(c.Redis.Interface.Config),
		// comic
		upComic: c.Host.Manga + _upComic,
		// contribute cache
		expireContribute: int32(time.Duration(c.Redis.Interface.ExpireContribute) / time.Second),
	}
	var err error
	if d.upClient, err = upgrpc.NewClient(c.UpArcClient); err != nil {
		panic(err)
	}
	if d.artClient, err = artclient.NewClient(c.ArtClient); err != nil {
		panic(err)
	}
	return
}

// UpArticles get article data from api.
func (d *Dao) UpArticles(c context.Context, mid int64, pn, ps int) (arts []*article.Meta, count int, err error) {
	arg := &artclient.UpArtMetasReq{Mid: mid, Pn: int32(pn), Ps: int32(ps)}
	res, err := d.artClient.UpArtMetas(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	arts = res.Articles
	count = int(res.Count)
	return
}

// UpArcs get upper archives up service.
func (d *Dao) UpArcs(ctx context.Context, mid, pn, ps int64, isCooperation bool) ([]*upgrpc.Arc, error) {
	var without []upgrpc.Without
	if !isCooperation {
		without = append(without, upgrpc.Without_staff)
	}
	reply, err := d.upClient.ArcPassed(ctx, &upgrpc.ArcPassedReq{Mid: mid, Pn: pn, Ps: ps, Without: without})
	if err != nil {
		if ecode.EqualError(ecode.NothingFound, err) {
			return nil, nil
		}
		return nil, err
	}
	return reply.Archives, nil
}

func (d *Dao) UpComics(c context.Context, mid int64, pn, ps int) (comics []*space.Comic, err error) {
	type params struct {
		UID      string `json:"uid"`
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
	}
	p := &params{
		UID:      strconv.FormatInt(mid, 10),
		Page:     pn,
		PageSize: ps,
	}
	bs, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", d.upComic, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *space.Comics `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.upComic)
		return
	}
	if res.Data != nil {
		comics = res.Data.ComicList
	}
	return
}
