// Package common
// 对接内容安全库
package common

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	api "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	arcGrpc "go-gateway/app/app-svr/archive/service/api"
)

const (
	_noHot       = "24" // 热门禁止
	_noRank      = "49" // 排行禁止
	_noIndex     = "50" // 分区动态禁止
	_noSearch    = "53" // 搜索禁止
	_noRecommend = "55" // 推荐禁止
	_noPushBlog  = "57" // 粉丝动态禁止

	_attrV2OnlySelf = 17 // 仅自己可见bit位
	_attrYes        = 1
)

// HitSixLimit 命中稿件六限， securityResp attrV2 可不传
// 逻辑：
// 1、【线上稿件露出风险】f_norecommend=1
// 2、【政策监管需要】(稿件自见=1) AND (f_nohot=1 OR f_norank=1 OR f_noindex=1 OR f_norecommend=1 OR f_nosearch=1 OR f_push_blog=1)
func (s *Service) HitSixLimit(ctx context.Context, aid int64) bool {
	if aid <= 0 {
		return true
	}

	var (
		securityResp []*api.InfoItem
		arc          *arcGrpc.SimpleArc
	)

	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var localErr error
		if securityResp, localErr = s.archiveDao.FlowControlInfoV2(ctx, aid, s.c.FlowControlAll); localErr != nil {
			log.Errorc(ctx, "HitSixLimit s.archiveDao.FlowControlInfoV2 err:%+v, aid:%d", localErr, aid)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var localErr error
		if arc, localErr = s.archiveDao.SimpleArc(ctx, aid); localErr != nil {
			log.Errorc(ctx, "HitSixLimit s.archiveDao.SimpleArc err:%+v, aid:%d", localErr, aid)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "HitSixLimit errgroup err:%+v", err)
		return true
	}

	// 查询管控信息失败，则降级为透出稿件
	if securityResp == nil || arc == nil {
		return false
	}
	return s.sixLimit(securityResp, arc.AttributeV2)
}

// HitSixLimitBatch 命中稿件六限批量接口
func (s *Service) HitSixLimitBatch(ctx context.Context, aids []int64) map[int64]bool {
	var (
		securityInfos map[int64]*api.FlowCtlInfoV2Reply
		arcs          map[int64]*arcGrpc.SimpleArc
		res           = make(map[int64]bool)
	)

	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var localErr error
		if securityInfos, localErr = s.archiveDao.FlowControlInfosV2(ctx, aids, s.c.FlowControlAll); localErr != nil {
			log.Errorc(ctx, "HitSixLimitBatch s.archiveDao.FlowControlInfoV2 err:%+v, aids:%+v", localErr, aids)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var localErr error
		arcs, localErr = s.archiveDao.SimpleArcs(ctx, aids)
		if localErr != nil {
			log.Errorc(ctx, "HitSixLimitBatch s.archiveDao.SimpleArc err:%+v, aids:%+v", localErr, aids)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "HitSixLimitBatch errgroup err:%+v", err)
		return res
	}

	for _, aid := range aids {
		// 查询管控信息失败，则降级为透出稿件
		if len(arcs) == 0 || len(securityInfos) == 0 || arcs[aid] == nil || securityInfos[aid] == nil {
			continue
		}
		res[aid] = s.sixLimit(securityInfos[aid].Items, arcs[aid].AttributeV2)
	}
	return res
}

// SixLimitFilter 过滤六限稿件
func (s *Service) SixLimitFilter(ctx context.Context, aids []int64) []int64 {
	if len(aids) == 0 {
		return aids
	}
	hitMap := s.HitSixLimitBatch(ctx, aids)
	res := make([]int64, 0)
	for _, aid := range aids {
		if !hitMap[aid] {
			res = append(res, aid)
		} else {
			log.Warnc(ctx, "SixLimitFilter filter aid:%d", aid)
		}
	}
	return res
}

func getByBit(attrV2 int64, bit uint) int32 {
	return int32((attrV2 >> bit) & int64(1))
}

func (s *Service) sixLimit(securityResp []*api.InfoItem, attrV2 int64) bool {
	// 查询管控信息失败，则降级为透出稿件
	if len(securityResp) == 0 {
		return false
	}

	securityMap := make(map[string]bool)
	for _, item := range securityResp {
		securityMap[item.Key] = item.Value == _attrYes
	}

	if securityMap[_noRecommend] {
		return true
	}
	if getByBit(attrV2, _attrV2OnlySelf) == _attrYes &&
		(securityMap[_noHot] || securityMap[_noRank] || securityMap[_noIndex] || securityMap[_noSearch] || securityMap[_noPushBlog]) {
		return true
	}

	return false
}
