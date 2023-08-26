package service

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"strconv"
	"time"

	l "go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/model/match"
	"strings"
)

func (s *Service) lotteryActionRedisClean(ctx context.Context, msg *match.Message) (err error) {
	if s.c.Lottery.RedisClean != 1 {
		return nil
	}
	time.Sleep(200 * time.Millisecond)
	v := &l.Record{}

	table := msg.Table
	countSplit := strings.Split(table, "_")
	if len(countSplit) != 4 {
		log.Error("lotteryRedisClean split count(%d)", len(countSplit))
		return
	}
	if err := json.Unmarshal(msg.New, v); err != nil {
		log.Errorc(ctx, "lotteryRedisClean json.Unmarshal() msg.New:%s,error(%v)", string(msg.New), err)
		return err
	}
	sid, err := strconv.ParseInt(countSplit[3], 10, 64)
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean err(%v)", err)
		return
	}
	err = s.dao.DeleteLotteryActionLog(ctx, sid, v.Mid)
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean DeleteLotteryActionLog err(%v)", err)
	}
	err = s.dao.NewDeleteLotteryActionLog(ctx, sid, v.Mid)
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean NewDeleteLotteryActionLog err(%v)", err)
	}
	err = s.dao.DeleteCacheLotteryTimes(ctx, sid, v.Mid, "used")
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean DeleteLotteryActionLog err(%v)", err)
	}
	err = s.dao.DeleteOldCacheLotteryTimes(ctx, sid, v.Mid, "used")
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean DeleteOldCacheLotteryTimes err(%v)", err)
	}
	err = s.dao.NewDeleteCacheLotteryTimes(ctx, sid, v.Mid, "used")
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean NewDeleteLotteryActionLog err(%v)", err)
	}
	if err == nil {
		log.Infoc(ctx, "lotteryRedisClean mid(%d) sid(%d) success", v.Mid, sid)
	}
	return err
}

func (s *Service) lotteryAddRedisClean(ctx context.Context, msg *match.Message) (err error) {
	if s.c.Lottery.RedisClean != 1 {
		return nil
	}
	time.Sleep(200 * time.Millisecond)
	v := &l.Record{}
	table := msg.Table
	countSplit := strings.Split(table, "_")
	if len(countSplit) != 4 {
		log.Error("lotteryRedisClean split count(%d)", len(countSplit))
		return
	}
	if err := json.Unmarshal(msg.New, v); err != nil {
		log.Errorc(ctx, "lotteryRedisClean json.Unmarshal() msg.New:%s,error(%v)", string(msg.New), err)
		return err
	}
	sid, err := strconv.ParseInt(countSplit[3], 10, 64)
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean err(%v)", err)
		return
	}
	err = s.dao.DeleteCacheLotteryTimes(ctx, sid, v.Mid, "add")
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean DeleteLotteryActionLog err(%v)", err)
	}
	err = s.dao.DeleteOldCacheLotteryTimes(ctx, sid, v.Mid, "add")
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean DeleteOldCacheLotteryTimes err(%v)", err)
	}
	err = s.dao.NewDeleteCacheLotteryTimes(ctx, sid, v.Mid, "add")
	if err != nil {
		log.Errorc(ctx, "lotteryRedisClean NewDeleteLotteryActionLog err(%v)", err)
	}
	if err == nil {
		log.Infoc(ctx, "lotteryRedisClean mid(%d) sid(%d) success", v.Mid, sid)
	}
	return err
}
