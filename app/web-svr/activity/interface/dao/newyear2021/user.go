package newyear2021

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/component"

	relationGrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"

	grab "git.bilibili.co/bapis/bapis-go/garb/service"
)

//TABLE STRUCTURE:
//create table if not exists new_year_user (`mid` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
//`lottery` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '奖券数',
//`exp` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '经验值',
//`ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
//`mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
//KEY `ix_mtime` (`mtime`));

const (
	cacheKey4UserPaid   = "bnj2021:paid:%v"
	cacheKey4UserFollow = "bnj2021:follow:%v:%v"
	userPaid            = 1
	userNotPaid         = 0

	userFollowed = 1
)

func (d *Dao) IsUserPaid(ctx context.Context, mid, suitId int64) (paid bool, err error) {
	shouldUpdateCache := false
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		if err != nil {
			log.Errorc(ctx, "bnj2021 IsUserPaid error: %v", err)
		}
		if shouldUpdateCache {
			ttl := 180 //default ttl 3min
			status := userNotPaid
			if paid {
				status = userPaid
				ttl = 3600 * 24 * 45 //user is paid, and will never change back to unpaid. so make ttl longer.
			}
			if err := conn.Send("SETEX", fmt.Sprintf(cacheKey4UserPaid, mid), ttl, status); err != nil {
				log.Errorc(ctx, "bnj2021 IsUserPaid update cache error: %v", err)
			}
		}
		_ = conn.Close()
	}()
	s, err := redis.Int(conn.Do("GET", fmt.Sprintf(cacheKey4UserPaid, mid)))
	if err == nil {
		paid = s == userPaid
		return
	}
	res, err := client.GarbClient.BNJUserFanNumBought(ctx, &grab.BNJUserFanNumBoughtReq{
		Mid:    mid,
		SuitID: suitId})
	if err != nil {
		return
	}
	paid = res.Bought
	shouldUpdateCache = true
	return
}

// isUserFollowed: 检查mid用户是否关注了fid用户
// 未关注时,不增加缓存
// 0--未关注
// 1--已关注
func (d *Dao) IsUserFollowed(ctx context.Context, mid, fid int64) (isFollowed bool, err error) {
	conn := component.GlobalBnjCache.Get(ctx)
	defer func() {
		if err != nil {
			log.Errorc(ctx, "bnj2021 IsUserFollowed error: %v", err)
		}
		if isFollowed {
			ttl := 180 //default ttl 3min
			status := userFollowed
			if err := conn.Send("SETEX", fmt.Sprintf(cacheKey4UserFollow, mid, fid), ttl, status); err != nil {
				log.Errorc(ctx, "bnj2021 IsUserFollowed update cache error: %v", err)
			}
		}
		_ = conn.Close()
	}()
	s, err := redis.Int(conn.Do("GET", fmt.Sprintf(cacheKey4UserFollow, mid, fid)))
	if err == nil {
		isFollowed = s == userFollowed
		return
	}
	resp, err := client.RelationClient.Relation(ctx, &relationGrpc.RelationReq{
		Mid: mid,
		Fid: fid,
	})
	if err != nil {
		log.Errorc(ctx, "call s.relGRPC.Relation error: %v", err)
		return
	}
	// Attribute: 1-悄悄关注 2-关注  6-好友 128-拉黑
	// Special:  0-不是特别关注 1-特别关注
	if resp.Attribute == 2 || resp.Attribute == 6 {
		isFollowed = true
	}
	return
}

func (d *Dao) IsUserFollowedStr(ctx context.Context, mid, fid int64) (isFollowed string) {
	isFollowed = "0"
	b, err := d.IsUserFollowed(ctx, mid, fid)
	if err == nil && b {
		isFollowed = "1"
	}
	return
}
