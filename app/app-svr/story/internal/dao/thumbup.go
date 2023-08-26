package dao

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

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

// HasLike user has like
func (d *dao) HasLike(c context.Context, buvid string, mid int64, messageIDs []int64) (res map[int64]int8, err error) {
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
func (d *dao) UserLikedCounts(c context.Context, mids []int64) (upCounts map[int64]int64, err error) {
	const _max = 20
	var (
		mutex = sync.Mutex{}
	)
	upCounts = make(map[int64]int64)
	g := errgroup.WithContext(c)
	if likeLen := len(mids); likeLen > 0 {
		for i := 0; i < likeLen; i += _max {
			var partMids []int64
			if i+_max > likeLen {
				partMids = mids[i:]
			} else {
				partMids = mids[i : i+_max]
			}
			g.Go(func(ctx context.Context) (err error) {
				result, err := d.userLikedCounts(ctx, partMids)
				if err != nil {
					log.Error("Failed to request single userLikedCounts: %+v", err)
					return
				}
				if len(result) > 0 {
					mutex.Lock()
					for k, v := range result {
						upCounts[k] = v
					}
					mutex.Unlock()
				}
				return
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (d *dao) userLikedCounts(c context.Context, mids []int64) (upCounts map[int64]int64, err error) {
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

func (d *dao) MultiLikeAnimation(ctx context.Context, aids []int64) (map[int64]*api.LikeAnimation, error) {
	args := &api.MultiLikeAnimationReq{
		Business:   "archive",
		MessageIds: aids,
	}
	reply, err := d.thumbupClient.MultiLikeAnimation(ctx, args)
	if err != nil {
		return nil, err
	}
	return reply.LikeAnimation, nil
}
