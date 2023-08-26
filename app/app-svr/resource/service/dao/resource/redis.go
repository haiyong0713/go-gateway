package resource

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
)

func keyDefaultPage(resourceId int64) string {
	return fmt.Sprintf("frontpage_default_%d", resourceId)
}

func keyOnlinePage(resourceId int64) string {
	return fmt.Sprintf("frontpage_online_%d", resourceId)
}

func keyHiddenPage(resourceId int64) string {
	return fmt.Sprintf("frontpage_hidden_%d", resourceId)
}

func (d *Dao) CacheDefaultPage(c context.Context, req *pb.FrontPageReq) (res *pb.FrontPage, err error) {
	var (
		bs   []byte
		key  = keyDefaultPage(req.ResourceId)
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("CacheDefaultPage conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = &pb.FrontPage{}
	if err = res.Unmarshal(bs); err != nil {
		log.Error("CacheDefaultPage json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheFrontPage
func (d *Dao) AddCacheDefaultPage(c context.Context, req *pb.FrontPageReq, res *pb.FrontPage) (err error) {
	var (
		key  = keyDefaultPage(req.ResourceId)
		conn = d.redis.Conn(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = res.Marshal(); err != nil {
		log.Error("AddCacheTimeLine json.Marshal req(%+v) error(%v)", res, err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.redisFrontPageExpire, bs); err != nil {
		log.Error("AddCacheTimeLine conn.Do SETEX key(%s) req(%+v) err(%+v)", key, string(bs), err)
		return
	}
	return
}

func (d *Dao) CacheHiddenPage(c context.Context, req *pb.FrontPageReq) (res []*pb.FrontPage, err error) {
	var (
		bs   []byte
		key  = keyHiddenPage(req.ResourceId)
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("CacheHiddenPage conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	pageResp := &pb.FrontPageResp{}
	if err = pageResp.Unmarshal(bs); err != nil {
		log.Error("CacheHiddenPage json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	res = pageResp.Hidden
	return
}

// AddCacheHiddenFrontPage
func (d *Dao) AddCacheHiddenPage(c context.Context, req *pb.FrontPageReq, res []*pb.FrontPage) (err error) {
	var (
		key  = keyHiddenPage(req.ResourceId)
		conn = d.redis.Conn(c)
		bs   []byte
	)
	defer conn.Close()
	resCache := &pb.FrontPageResp{}
	resCache.Hidden = res
	if bs, err = resCache.Marshal(); err != nil {
		log.Error("AddCacheHiddenPage json.Marshal req(%+v) error(%v)", res, err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.redisFrontPageExpire, bs); err != nil {
		log.Error("AddCacheHiddenPage conn.Do SETEX key(%s) req(%+v) err(%+v)", key, string(bs), err)
		return
	}
	return
}

func (d *Dao) CacheOnlinePage(c context.Context, req *pb.FrontPageReq) (res []*pb.FrontPage, err error) {
	var (
		bs   []byte
		key  = keyOnlinePage(req.ResourceId)
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("CacheOnlinePage conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	pageResp := &pb.FrontPageResp{}
	if err = pageResp.Unmarshal(bs); err != nil {
		log.Error("CacheOnlinePage json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	res = pageResp.Online
	return
}

// AddCacheOnlineFrontPage
func (d *Dao) AddCacheOnlinePage(c context.Context, req *pb.FrontPageReq, res []*pb.FrontPage) (err error) {
	var (
		key  = keyOnlinePage(req.ResourceId)
		conn = d.redis.Conn(c)
		bs   []byte
	)
	defer conn.Close()
	resCache := &pb.FrontPageResp{}
	resCache.Online = res
	if bs, err = resCache.Marshal(); err != nil {
		log.Error("AddCacheOnlinePage json.Marshal req(%+v) error(%v)", res, err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.redisFrontPageExpire, bs); err != nil {
		log.Error("AddCacheOnlinePage conn.Do SETEX key(%s) req(%+v) err(%+v)", key, string(bs), err)
		return
	}
	return
}
