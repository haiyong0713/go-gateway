package search

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"strings"
	"sync"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/time"
	pb "go-gateway/app/app-svr/app-feed/admin/api/search"
	model "go-gateway/app/app-svr/app-feed/admin/model/search"

	accountGRPC "git.bilibili.co/bapis/bapis-go/account/service"
	relationGRPC "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

// 新增up主别名
func (s *Service) AddUpAlias(c *bm.Context, req *pb.AddUpAliasReq) error {
	accReq := &accountGRPC.MidReq{
		Mid: req.Mid,
	}
	var (
		accRep *accountGRPC.InfoReply
		err    error
	)
	if accRep, err = s.accClient.Info3(c, accReq); err != nil {
		return ecode.Error(-400, "请输入正确的UID: "+err.Error())
	}

	param := &model.UpAlias{
		Mid:         req.Mid,
		Nickname:    accRep.Info.Name,
		SearchWords: req.SearchWords,
		Stime:       time.Time(req.Stime),
		Etime:       time.Time(req.Etime),
		IsForever:   req.IsForever,
		Applier:     req.Applier,
	}

	if err = s.dao.AddUpAlias(param); err != nil {
		if strings.HasPrefix(err.Error(), "Error 1062") {
			return ecode.Error(-400, "该UID已配置别名，请返回列表修改原有配置")
		}
	}

	return err
}

// 编辑up主别名
func (s *Service) EditUpAlias(_ *bm.Context, req *pb.EditUpAliasReq) error {
	param := &model.UpAlias{
		Id:          req.Id,
		SearchWords: req.SearchWords,
		Stime:       time.Time(req.Stime),
		Etime:       time.Time(req.Etime),
		IsForever:   req.IsForever,
		Applier:     req.Applier,
	}
	return s.dao.EditUpAlias(param)
}

// 上下线up主别名
func (s *Service) ToggleUpAlias(_ *bm.Context, req *pb.ToggleUpAliasReq) error {
	return s.dao.ToggleAlias(req.Id, req.State)
}

// 查找up主别名
func (s *Service) SearchUpAlias(c *bm.Context, req *pb.SearchUpAliasReq) (resp *pb.SearchUpAliasRep, err error) {
	resp = new(pb.SearchUpAliasRep)
	resp.Items = make([]*pb.UpAlias, 0)
	resp.Pager = &pb.PageInfo{
		Num:   req.Pn,
		Size_: req.Ps,
		Total: 0,
	}
	var (
		items []*model.UpAlias
		total int32

		mids = make([]int64, 0)
	)
	if items, total, err = s.dao.FindAliasByParam(req.Mid, req.Nickname, req.SearchWords, req.Applier, req.Pn, req.Ps); err != nil {
		return resp, err
	}
	resp.Pager.Total = total

	if len(items) == 0 {
		return resp, nil
	}
	for _, ele := range items {
		mids = append(mids, ele.Mid)
		resp.Items = append(resp.Items, ele.GetEntityForPB())
	}

	eg := errgroup.WithContext(c)
	lock := sync.Mutex{}

	// 查询粉丝数
	eg.Go(func(ctx context.Context) error {
		var (
			peakRep *relationGRPC.StatsReply
			e       error
		)
		peakReq := &relationGRPC.MidsReq{
			Mids: mids,
		}
		if peakRep, e = s.relationClient.PeakStats(c, peakReq); e != nil {
			log.Error("s.relationClient.PeakStats err: %s", e)
			return e
		}

		for i := range resp.Items {
			mid := resp.Items[i].Mid
			if v, ok := peakRep.StatReplyMap[mid]; ok {
				lock.Lock()
				resp.Items[i].FansCount = v.Follower
				lock.Unlock()
			}
		}
		return nil
	})

	// 查询封禁状态
	eg.Go(func(ctx context.Context) error {
		var (
			proRep *accountGRPC.ProfilesWithoutPrivacyReply
			e      error
		)
		infoReq := &accountGRPC.MidsReq{
			Mids: mids,
		}
		if proRep, e = s.accClient.ProfilesWithoutPrivacy3(c, infoReq); e != nil {
			log.Error("s.accClient.ProfilesWithoutPrivacy3 err: %s", e)
			return e
		}

		for i := range resp.Items {
			mid := resp.Items[i].Mid
			if v, ok := proRep.ProfilesWithoutPrivacy[mid]; ok {
				if v.Silence == 1 {
					lock.Lock()
					resp.Items[i].Nickname += "(已封禁)"
					lock.Unlock()
				}
			}
		}
		return nil
	})
	err = eg.Wait()

	return resp, err

}

// 上下线up主别名
func (s *Service) SyncUpAlias(_ *bm.Context, req *pb.SyncUpAliasReq) (resp *pb.SyncUpAliasRep, err error) {
	resp = new(pb.SyncUpAliasRep)
	resp.Items = make([]*pb.SyncUpAlias, 0)

	var raw []*model.UpAlias
	if raw, err = s.dao.FindAliasForSync(req.EffectTime); err != nil {
		return resp, err
	}

	for _, ele := range raw {
		resp.Items = append(resp.Items, ele.GetEntityForSyncPB())
	}

	return resp, nil
}

// 导出up主别名信息
func (s *Service) ExportUpAlias(_ *bm.Context) (string, error) {
	var (
		items []*model.UpAlias
		err   error
	)
	if items, err = s.dao.FindAllAlias(); err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", nil
	}
	result := make([]string, 0)
	result = append(result, fmt.Sprintf(
		"%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
		"ID", "MID", "Nickname", "SearchWords", "STime", "ETime", "IsForever", "Applier", "State", "CTime",
	))

	for _, ele := range items {
		result = append(result, ele.GetEntityForExport())
	}
	return strings.Join(result, "\n"), nil
}
