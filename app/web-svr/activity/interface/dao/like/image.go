package like

import (
	"context"
	"fmt"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"go-common/library/cache/redis"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

func imageUserKey(sid int64, day string, typ int) string {
	return fmt.Sprintf("img_rk_%d_%s_%d", sid, day, typ)
}

func (d *Dao) ImageUserRankList(c context.Context, sid int64, day string, typ int, count int64) ([]*like.ImgUser, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := imageUserKey(sid, day, typ)
	values, err := redis.Values(conn.Do("ZREVRANGE", key, 0, count, "WITHSCORES"))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return nil, nil
		} else {
			err = errors.Wrapf(err, "conn.Do(ZREVRANGE) key:%s", key)
			return nil, err
		}
	}
	var res []*like.ImgUser
	for len(values) > 0 {
		r := &like.ImgUser{SimpleUser: &like.SimpleUser{}}
		if values, err = redis.Scan(values, &r.Mid, &r.ImageScore); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

func (d *Dao) ImageUserRank(c context.Context, sid, mid int64, day string, typ int) (rank int64, score float64, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := imageUserKey(sid, day, typ)
	if err = conn.Send("ZREVRANK", key, mid); err != nil {
		err = errors.Wrapf(err, "ImageUserRank conn.Send(ZREVRANK) key:%s", key)
		return
	}
	if err = conn.Send("ZSCORE", key, mid); err != nil {
		err = errors.Wrapf(err, "ImageUserRank conn.Send(ZSCORE) key:%s", key)
		return
	}
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, "ImageUserRank conn.Flush()")
		return
	}
	if rank, err = redis.Int64(conn.Receive()); err != nil {
		rank = bwsmdl.DefaultRank
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrapf(err, "ImageUserRank redis.Int64() key:%s", key)
		}
		return
	}
	if score, err = redis.Float64(conn.Receive()); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrapf(err, "ImageUserRank redis.Float64() key:%s", key)
		}
	}
	return
}
