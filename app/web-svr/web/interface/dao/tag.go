package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/web/interface/model"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"github.com/pkg/errors"
)

const (
	_redisTagAv = "t_a_"
	_tagFeedURL = "/feed/tag/top"
)

func tagAidKey(tid int64) string {
	return _redisTagAv + strconv.FormatInt(tid, 10)
}

// TagAids provides aids via tag
func (d *Dao) TagAids(c context.Context, tid int64) (res []int64, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("tag", strconv.FormatInt(tid, 10))
	params.Set("pn", "1")
	params.Set("rn", strconv.Itoa(d.c.Tag.PageSize))
	params.Set("src", "1") // plat. PC:1, APP:2
	rs := &model.TagAids{}
	if err = d.httpR.Get(c, d.c.Host.Data+_tagFeedURL, ip, params, rs); err != nil {
		log.Error("tag d.httpR.Get(%s, %s, %v) error(%v)", d.c.Host.Data+_tagFeedURL, ip, params, err)
		return
	}
	if rs.Code != ecode.OK.Code() {
		err = ecode.Int(rs.Code)
		return
	}
	res = rs.Data
	return
}

// TagAidsBakCache gets avids cache
func (d *Dao) TagAidsBakCache(c context.Context, tid int64) (res []int64, err error) {
	var (
		conn = d.redisBak.Get(c)
		key  = tagAidKey(tid)
		s    string
	)
	defer conn.Close()
	if s, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET, %s) error(%v)", key, err)
		}
		return
	}
	if res, err = xstr.SplitInts(s); err != nil {
		log.Error("xstr.SplitInts(%s) error(%v)", s, err)
	}
	return
}

// SetTagAidsBakCache set the avids cache
func (d *Dao) SetTagAidsBakCache(c context.Context, tid int64, aids []int64) (err error) {
	var (
		conn = d.redisBak.Get(c)
		key  = tagAidKey(tid)
	)
	defer conn.Close()
	s := xstr.JoinInts(aids)
	if err = conn.Send("SET", key, s); err != nil {
		log.Error("conn.Do(SET, %s, %s) error(%v)", key, s, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.redisTagBakExpire); err != nil {
		log.Error("conn.Send(EXPIRE, %s, %d) error(%v)", key, d.redisTagBakExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) AddSub(c context.Context, mid int64, tagIDs []int64) error {
	req := &taggrpc.AddSubReq{Mid: mid, Tids: tagIDs}
	if _, err := d.tagClient.AddSub(c, req); err != nil {
		err = errors.Wrapf(err, "d.tagClient.AddSub(%+v)", req)
		return err
	}
	return nil
}

func (d *Dao) CancelSub(c context.Context, mid, tagID int64) error {
	req := &taggrpc.CancelSubReq{Mid: mid, Tid: tagID}
	if _, err := d.tagClient.CancelSub(c, req); err != nil {
		err = errors.Wrapf(err, "d.tagClient.CancelSub(%+v)", req)
		return err
	}
	return nil
}

func (d *Dao) RidsByTag(c context.Context, req *taggrpc.RidsByTagReq) (*taggrpc.RidsByTagReply, error) {
	var (
		err   error
		reply *taggrpc.RidsByTagReply
	)
	if reply, err = d.tagClient.RidsByTag(c, req); err != nil {
		err = errors.Wrapf(err, "d.tagClient.RidsByTag(%+v)", req)
		return nil, err
	}
	return reply, nil
}
