package upper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache/credis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	feedApi "git.bilibili.co/bapis/bapis-go/community/service/feed"
	feedArtApi "git.bilibili.co/bapis/bapis-go/community/service/feed/article"

	"github.com/pkg/errors"
)

// Dao is feed dao.
type Dao struct {
	// rpc
	feedRPC feedApi.FeedClient
	// redis
	redis     credis.Redis
	expireRds int32
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// redis init
		redis:     credis.NewRedis(c.Redis.Upper.Config),
		expireRds: int32(time.Duration(c.Redis.Upper.ExpireUpper) / time.Second),
	}
	var err error
	if d.feedRPC, err = feedApi.NewClient(c.FeedRPC); err != nil {
		panic(fmt.Sprintf("accountgrpc NewClientt error (%+v)", err))
	}
	return
}

// Ping check redis connection
func (d *Dao) Ping(c context.Context) (err error) {
	conn := d.redis.Conn(c)
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}

func (d *Dao) Feed(c context.Context, mid int64, pn, ps int) ([]*feedApi.Record, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &feedApi.FeedReq{Mid: mid, Pn: int64(pn), Ps: int64(ps), RealIP: ip}
	fs, err := d.feedRPC.AppFeed(c, arg)
	if err != nil {
		if err == ecode.NothingFound {
			return nil, nil
		}
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	return fs.GetRecords(), nil
}

func (d *Dao) ArchiveFeed(c context.Context, mid int64, pn, ps int) ([]*feedApi.Record, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &feedApi.FeedReq{Mid: mid, Pn: int64(pn), Ps: int64(ps), RealIP: ip}
	fs, err := d.feedRPC.ArchiveFeed(c, arg)
	if err != nil {
		if err == ecode.NothingFound {
			return nil, nil
		}
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	b, _ := json.Marshal(&fs)
	log.Info("ArchiveFeed mid(%d) list(%s)", mid, b)
	return fs.GetRecords(), nil
}

func (d *Dao) AppUnreadCount(c context.Context, mid int64, withoutBangumi bool) (int, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &feedApi.AppUnreadCountReq{Mid: mid, WithoutBangumi: withoutBangumi, RealIP: ip}
	res, err := d.feedRPC.AppUnreadCount(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return 0, err
	}
	return int(res.GetCount()), nil
}

func (d *Dao) ArticleFeed(c context.Context, mid int64, pn, ps int) ([]*feedArtApi.Meta, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &feedApi.FeedReq{Mid: mid, Pn: int64(pn), Ps: int64(ps), RealIP: ip}
	fs, err := d.feedRPC.ArticleFeed(c, arg)
	if err != nil {
		if err == ecode.NothingFound {
			return nil, nil
		}
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	return fs.GetMeta(), nil
}

func (d *Dao) ArticleUnreadCount(c context.Context, mid int64) (int, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &feedApi.ArticleUnreadCountReq{Mid: mid, RealIP: ip}
	res, err := d.feedRPC.ArticleUnreadCount(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return 0, err
	}
	return int(res.GetCount()), nil
}
