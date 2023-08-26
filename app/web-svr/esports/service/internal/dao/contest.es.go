package dao

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (d *dao) GetContestsCacheOrEs(ctx context.Context, contestsQueryParams *model.ContestsQueryParamsModel) (contestIds []int64, total int, cache bool, err error) {
	contestIds, total, err = d.getEsCacheContestIds(ctx, contestsQueryParams)
	if err != nil && err != redis.ErrNil {
		log.Errorc(ctx, "[Dao][GetContestsCacheOrEs][getEsCacheContestIds][Error], err:%+v", err)
		return
	}
	if err == nil {
		cache = true
		return
	}
	contestIds, total, err = d.SearchContestsByTime(ctx, contestsQueryParams)
	if err != nil {
		log.Errorc(ctx, "[Dao][GetContestsCacheOrEs][SearchContestsByTime][Error], err:%+v", err)
		return
	}
	_ = d.setEsCacheContestIds(ctx, contestsQueryParams, contestIds, total)
	return
}

func (d *dao) getEsCacheContestIds(ctx context.Context, req *model.ContestsQueryParamsModel) (contestIds []int64, total int, err error) {
	md5Str, err := formatEsCacheKeyMd5(ctx, req)
	if err != nil {
		log.Errorc(ctx, "[getEsCacheContestIds][formatEsCacheKeyMd5][Error], err:%+v", err)
		return
	}
	contestIds, total, err = d.GetEsContestIdsCache(ctx, md5Str)
	if err != nil {
		log.Errorc(ctx, "[getEsCacheContestIds][GetEsContestIdsCache][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) setEsCacheContestIds(ctx context.Context, req *model.ContestsQueryParamsModel, contestIds []int64, total int) (err error) {
	md5Str, err := formatEsCacheKeyMd5(ctx, req)
	if err != nil {
		log.Errorc(ctx, "[setEsCacheContestIds][formatEsCacheKeyMd5][Error], err:%+v", err)
		return
	}
	err = d.SetEsContestIdsCache(ctx, md5Str, contestIds, total)
	if err != nil {
		log.Errorc(ctx, "[setEsCacheContestIds][SetEsContestIdsCache][Error], err:%+v", err)
		return
	}
	return
}

func formatEsCacheKeyMd5(ctx context.Context, req *model.ContestsQueryParamsModel) (md5Str string, err error) {
	bytes, err := json.Marshal(req)
	if err != nil {
		log.Errorc(ctx, "[getEsCacheContestIds][Error], err:%+v", err)
		err = xecode.Errorf(xecode.ServerErr, "内部错误")
		return
	}
	h := md5.New()
	h.Write(bytes)
	newStr := h.Sum(nil)
	md5Str = fmt.Sprintf("%X", newStr)
	return
}
