package dao

import (
	"context"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	"go-common/library/cache/redis"
	xecode "go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	"go-gateway/app/app-svr/archive-push/ecode"
)

func (d *Dao) GetArcsByBVIDs(bvids []string) (res map[int64]*archiveGRPC.Arc, err error) {
	res = make(map[int64]*archiveGRPC.Arc)
	if len(bvids) == 0 {
		return
	}
	avids := make([]int64, 0)
	var avid int64
	for _, bvid := range bvids {
		if avid, err = util.BvToAv(bvid); err != nil {
			log.Error("Dao: GetArcsByBVIDs Error (%v)", err)
			err = ecode.AVBVIDConvertingError
			continue
		}
		avids = append(avids, avid)
	}
	req := &archiveGRPC.ArcsRequest{Aids: avids}
	var reply *archiveGRPC.ArcsReply
	if reply, err = d.archiveGRPCClient.Arcs(context.Background(), req); err != nil {
		return nil, err
	}
	res = reply.Arcs
	return
}

func (d *Dao) GetArcByAID(aid int64) (res *archiveGRPC.Arc, err error) {
	if aid == 0 {
		return
	}
	res = &archiveGRPC.Arc{}
	req := &archiveGRPC.ArcRequest{Aid: aid}
	var reply *archiveGRPC.ArcReply
	if reply, err = d.archiveGRPCClient.Arc(context.Background(), req); err != nil {
		return nil, err
	}
	res = reply.Arc
	return
}

// GetUpArcsByMID 根据用户MID获取已通过的稿件
func (d *Dao) GetUpArcsByMID(mid int64) (res []*archiveGRPC.Arc, err error) {
	if mid == 0 {
		return
	}
	req := &archiveGRPC.UpArcsRequest{Mid: mid}
	var reply *archiveGRPC.UpArcsReply
	if reply, err = d.archiveGRPCClient.UpArcs(context.Background(), req); err != nil {
		return nil, err
	}
	res = reply.Arcs
	return
}

func (d *Dao) PutBVIDsForWhiteList(vendorID int64, bvids []string) (err error) {
	if len(bvids) == 0 {
		return
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	args := redis.Args{}
	if redisKey, _err := model.GetArchiveWhiteListKeyForVendor(vendorID); _err != nil {
		err = _err
		return
	} else {
		args = args.Add(redisKey)
	}
	args = args.AddFlat(bvids)
	if _, err = conn.Do("SADD", args...); err != nil {
		log.Error("Dao: PutBVIDsForWhiteList Error (%v)", err)
	}
	return
}

func (d *Dao) RemoveBVIDsForWhiteList(vendorID int64, bvids []string) (err error) {
	if len(bvids) == 0 {
		return
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	args := redis.Args{}
	if redisKey, _err := model.GetArchiveWhiteListKeyForVendor(vendorID); _err != nil {
		err = _err
		return
	} else {
		args = args.Add(redisKey)
	}
	args = args.AddFlat(bvids)
	if _, err = conn.Do("SREM", args...); err != nil {
		log.Error("Dao: RemoveBVIDsForWhiteList(%d, %v) Error (%v)", vendorID, bvids, err)
	}
	return
}

// GetBVIDsWhiteList 根据vendor查询所有白名单稿件
func (d *Dao) GetBVIDsWhiteList(vendorID int64) (bvids []string, err error) {
	args := redis.Args{}
	if redisKey, _err := model.GetArchiveWhiteListKeyForVendor(vendorID); _err != nil {
		err = _err
		return
	} else {
		args = args.Add(redisKey)
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	if bvids, err = redis.Strings(conn.Do("SMEMBERS", args...)); err != nil {
		log.Error("Dao: GetBVIDsWhiteList Error (%v)", err)
	}
	return
}

// GetAuthorBVIDsWhiteList 根据vendor和mid获取作者的白名单稿件
func (d *Dao) GetAuthorBVIDsWhiteList(vendorID int64, mid int64) (bvids []string, err error) {
	bvids = make([]string, 0)
	args := redis.Args{}
	if redisKey, _err := model.GetAuthorWhiteListKeyByAuthor(vendorID, mid); _err != nil {
		err = _err
		return
	} else {
		args = args.Add(redisKey)
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	if _bvids, _err := redis.Strings(conn.Do("SMEMBERS", args...)); _err != nil {
		log.Error("Dao: GetAuthorBVIDsWhiteList Error (%v)", _err)
	} else if len(_bvids) <= 1 {
		return
	} else {
		for _, bvid := range _bvids {
			if bvid != "true" && bvid != "" {
				bvids = append(bvids, bvid)
			}
		}
	}
	return
}

// PutAuthorBVIDsForWhiteList 将稿件添加进作者的稿件白名单
func (d *Dao) PutAuthorBVIDsForWhiteList(vendorID int64, mid int64, bvids []string) (err error) {
	if vendorID == 0 || mid == 0 || len(bvids) == 0 {
		return xecode.RequestErr
	}
	args := redis.Args{}
	if redisKey, _err := model.GetAuthorWhiteListKeyByAuthor(vendorID, mid); _err != nil {
		err = _err
		return
	} else {
		args = args.Add(redisKey)
	}
	for _, bvid := range bvids {
		_bvid := bvid
		args = args.Add(_bvid)
	}
	conn := d.redis.Conn(context.Background())
	defer conn.Close()
	if _, _err := conn.Do("SADD", args...); _err != nil {
		log.Error("Dao: AddAuthorBVIDsWhiteList Error (%v)", _err)
	}
	return
}
