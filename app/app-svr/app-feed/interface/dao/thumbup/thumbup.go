package thumbup

import (
	"context"

	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"github.com/pkg/errors"
)

const (
	// 用户获取点赞的业务
	_businessLike = "archive"
	_articleLike  = "article"
	_dynamicLike  = "dynamic"
	_albumLike    = "album"
	_clipLike     = "clip"
	_cheeseLike   = "cheese"
)

// Dao is tag dao
type Dao struct {
	thumbupClient api.ThumbupClient
}

// New initial tag dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.thumbupClient, err = api.NewClient(c.ThumbupGRPC); err != nil {
		panic(err)
	}
	return
}

// HasLike user has like
func (d *Dao) HasLike(c context.Context, buvid string, mid int64, messageIDs []int64) (res map[int64]int8, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	if mid > 0 {
		arg := &api.HasLikeReq{
			Business:   _businessLike,
			MessageIds: messageIDs,
			Mid:        mid,
			IP:         ip,
		}
		reply, err := d.thumbupClient.HasLike(c, arg)
		if err != nil {
			err = errors.Wrapf(err, "%v", arg)
			return nil, err
		}
		res = make(map[int64]int8)
		for k, v := range reply.States {
			res[k] = int8(v.State)
		}
	} else {
		arg := &api.BuvidHasLikeReq{
			Business:   _businessLike,
			MessageIds: messageIDs,
			Buvid:      buvid,
			IP:         ip,
		}
		reply, err := d.thumbupClient.BuvidHasLike(c, arg)
		if err != nil {
			err = errors.Wrapf(err, "%v", arg)
			return nil, err
		}
		res = make(map[int64]int8)
		for k, v := range reply.States {
			res[k] = int8(v.State)
		}
	}

	return
}

// UserLikedCounts 获取用户总点赞数
func (d *Dao) UserLikedCounts(c context.Context, mids []int64) (upCounts map[int64]int64, err error) {
	upCounts = map[int64]int64{}
	arg := &api.BatchLikedCountsReq{Mids: mids, Businesses: []string{_businessLike, _articleLike, _dynamicLike, _albumLike, _clipLike, _cheeseLike}}
	reply, err := d.thumbupClient.BatchLikedCounts(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if reply == nil {
		return
	}
	for mid, v := range reply.LikeCounts {
		if v == nil {
			continue
		}
		var allNum int64
		for _, count := range v.Records {
			allNum += count
		}
		upCounts[mid] = allNum
	}
	return
}

func (d *Dao) GetLikeStates(ctx context.Context, aids []int64) (map[int64]*api.StatState, error) {
	arg := &api.StatsReq{
		Business:   _businessLike,
		MessageIds: aids,
		IP:         metadata.String(ctx, metadata.RemoteIP),
	}
	reply, err := d.thumbupClient.Stats(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.Stats, nil
}
