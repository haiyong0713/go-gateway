package teen_manual

import (
	"context"
	"sync"
	"time"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	familymdl "go-gateway/app/app-svr/app-feed/admin/model/family"
	membermdl "go-gateway/app/app-svr/app-feed/admin/model/member"
	spmodemdl "go-gateway/app/app-svr/app-feed/admin/model/spmode"
	model "go-gateway/app/app-svr/app-feed/admin/model/teen_manual"
)

func (s *Service) Search(ctx context.Context, req *model.SearchTeenManualReq) (*model.SearchTeenManualRly, error) {
	total, teenUsers, err := s.spmodeDao.PagingTeenManual(req.Mid, req.Operator, req.Pn, req.Ps)
	if err != nil {
		return nil, ecode.Error(ecode.ServerErr, "数据查询失败")
	}
	// 只筛选mid，可能还无此条记录
	if len(teenUsers) == 0 && req.Mid > 0 && req.Operator == "" {
		teenUsers = append(teenUsers, &spmodemdl.TeenagerUsers{
			Mid:   req.Mid,
			Model: spmodemdl.ModelTeenager,
		})
	}
	accounts, ageChecks := s.fetchSearchMaterial(ctx, extractMidFromTeenUsers(teenUsers))
	items := make([]*model.TeenManualItem, 0, len(teenUsers))
	for _, teen := range teenUsers {
		if teen == nil || teen.Mid <= 0 {
			continue
		}
		item := buildSearchItem(teen, accounts[teen.Mid], ageChecks[teen.Mid])
		items = append(items, item)
	}
	return &model.SearchTeenManualRly{
		Page: &model.Page{Num: req.Pn, Size: req.Ps, Total: total},
		List: items,
	}, nil
}

func (s *Service) fetchSearchMaterial(ctx context.Context, mids []int64) (map[int64]*accountgrpc.Info, map[int64]*membergrpc.RealnameTeenAgeCheckReply) {
	if len(mids) == 0 {
		return nil, nil
	}
	eg := errgroup.WithContext(ctx)
	// 账号信息
	var accounts map[int64]*accountgrpc.Info
	eg.Go(func(ctx context.Context) error {
		if infos, err := s.accountDao.Infos3(ctx, mids); err == nil {
			accounts = infos
		}
		return nil
	})
	// 实名认证信息
	ageChecks := make(map[int64]*membergrpc.RealnameTeenAgeCheckReply, len(mids))
	lock := sync.Mutex{}
	for _, v := range mids {
		mid := v
		eg.Go(func(ctx context.Context) error {
			ageCheck, err := s.memberDao.RealnameTeenAgeCheck(ctx, mid)
			if err != nil || ageCheck == nil {
				return nil
			}
			lock.Lock()
			defer lock.Unlock()
			ageChecks[mid] = ageCheck
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("Fail to fetch SearchMaterial, mids=%+v error=%+v", mids, err)
		return nil, nil
	}
	return accounts, ageChecks
}

func (s *Service) Open(ctx context.Context, req *model.OpenReq, username string) error {
	teenUser, ageCheck, openFamily, err := s.fetchOpenMaterial(ctx, req.Mid)
	if err != nil {
		return err
	}
	if openFamily {
		return ecode.Error(ecode.RequestErr, "操作失败！用户已绑定亲子平台，不符合人工强拉条件")
	}
	if ageCheck.Realname == membermdl.RealnameVerified && ageCheck.After18 {
		return ecode.Error(ecode.RequestErr, "操作失败！用户实名年龄≥18，不符合人工强拉条件")
	}
	if teenUser != nil {
		if teenUser.ManualForce == spmodemdl.ManualForceOpen {
			return ecode.Error(ecode.RequestErr, "操作失败！当前uid强拉状态已开启，请刷新页面")
		}
		affected, err := s.spmodeDao.OpenManualForce(teenUser.ID, username)
		if err != nil {
			return ecode.Error(ecode.ServerErr, "人工强拉状态更新失败")
		}
		if affected == 0 {
			return ecode.Error(ecode.RequestErr, "操作失败！当前uid强拉状态已开启，请刷新页面")
		}
	} else {
		if err := s.spmodeDao.AddTeenUser(&spmodemdl.TeenagerUsers{
			Mid:         req.Mid,
			Model:       spmodemdl.ModelTeenager,
			ManualForce: spmodemdl.ManualForceOpen,
			MfOperator:  username,
			MfTime:      xtime.Time(time.Now().Unix()),
		}); err != nil {
			return ecode.Error(ecode.ServerErr, "人工强拉状态更新失败")
		}
	}
	_ = s.worker.Do(ctx, func(ctx context.Context) {
		_ = s.dao.AddManualLog(&model.TeenagerManualLog{
			Mid:      req.Mid,
			Operator: username,
			Content:  "人工强拉",
			Remark:   req.Remark,
		})
		if err := retry.WithAttempts(ctx, "del_cache_teen_user", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return s.spmodeDao.DelCacheModelUser(ctx, req.Mid)
		}); err != nil {
			log.Error("日志告警 删除青少年缓存失败, mid=%+v error=%+v", req.Mid, err)
		}
	})
	return nil
}

func (s *Service) fetchOpenMaterial(ctx context.Context, mid int64) (*spmodemdl.TeenagerUsers, *membergrpc.RealnameTeenAgeCheckReply, bool, error) {
	eg := errgroup.WithCancel(ctx)
	var teenUser *spmodemdl.TeenagerUsers
	eg.Go(func(ctx context.Context) error {
		rly, err := s.spmodeDao.TeenagerUserByMidModel(mid, spmodemdl.ModelTeenager)
		if err != nil {
			return ecode.Error(ecode.ServerErr, "数据查询失败")
		}
		teenUser = rly
		return nil
	})
	var ageCheck *membergrpc.RealnameTeenAgeCheckReply
	eg.Go(func(ctx context.Context) error {
		rly, err := s.memberDao.RealnameTeenAgeCheck(ctx, mid)
		if err != nil || rly == nil {
			return ecode.Error(ecode.ServerErr, "实名认证信息获取失败")
		}
		ageCheck = rly
		return nil
	})
	var parentRels []*familymdl.FamilyRelation
	eg.Go(func(ctx context.Context) error {
		rly, err := s.spmodeDao.FamilyRelsOfParent(mid)
		if err != nil {
			return ecode.Error(ecode.ServerErr, "亲子信息获取失败")
		}
		parentRels = rly
		return nil
	})
	var childRel *familymdl.FamilyRelation
	eg.Go(func(ctx context.Context) error {
		rly, err := s.spmodeDao.FamilyRelOfChild(mid)
		if err != nil {
			return ecode.Error(ecode.ServerErr, "亲子信息获取失败")
		}
		childRel = rly
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, false, err
	}
	var openFamily bool
	if len(parentRels) > 0 || childRel != nil {
		openFamily = true
	}
	return teenUser, ageCheck, openFamily, nil
}

func (s *Service) Quit(ctx context.Context, req *model.QuitReq, username string) error {
	teenUser, err := s.spmodeDao.TeenagerUserByMidModel(req.Mid, spmodemdl.ModelTeenager)
	if err != nil {
		return ecode.Error(ecode.ServerErr, "数据查询失败")
	}
	if teenUser == nil {
		return ecode.NothingFound
	}
	if teenUser.ManualForce == spmodemdl.ManualForceQuit {
		return ecode.Error(ecode.RequestErr, "操作失败！当前uid强拉状态已关闭，请刷新页面")
	}
	affected, err := s.spmodeDao.QuitManualForce(teenUser.ID, username)
	if err != nil {
		return ecode.Error(ecode.ServerErr, "人工强拉状态更新失败")
	}
	if affected == 0 {
		return ecode.Error(ecode.RequestErr, "操作失败！当前uid强拉状态已关闭，请刷新页面")
	}
	_ = s.worker.Do(ctx, func(ctx context.Context) {
		_ = s.dao.AddManualLog(&model.TeenagerManualLog{
			Mid:      req.Mid,
			Operator: username,
			Content:  "人工解除",
			Remark:   req.Remark,
		})
		if err := retry.WithAttempts(ctx, "del_cache_teen_user", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			return s.spmodeDao.DelCacheModelUser(ctx, req.Mid)
		}); err != nil {
			log.Error("日志告警 删除青少年缓存失败, mid=%+v error=%+v", req.Mid, err)
		}
	})
	return nil
}

func (s *Service) Log(ctx context.Context, req *model.TeenManualLogReq) (*model.TeenManualLogRly, error) {
	logs, err := s.dao.LogsByMid(req.Mid)
	if err != nil {
		return nil, ecode.Error(ecode.ServerErr, "数据查询失败")
	}
	return &model.TeenManualLogRly{List: logs}, nil
}

func extractMidFromTeenUsers(teenUsers []*spmodemdl.TeenagerUsers) []int64 {
	// teenUsers.mid已唯一
	mids := make([]int64, 0, len(teenUsers))
	for _, user := range teenUsers {
		if user == nil || user.Mid <= 0 {
			continue
		}
		mids = append(mids, user.Mid)
	}
	return mids
}

func buildSearchItem(teen *spmodemdl.TeenagerUsers, account *accountgrpc.Info, age *membergrpc.RealnameTeenAgeCheckReply) *model.TeenManualItem {
	item := &model.TeenManualItem{
		Mid:         teen.Mid,
		Name:        account.GetName(),
		State:       teen.State,
		ManualForce: teen.ManualForce,
		Operator:    teen.MfOperator,
		OperateTime: teen.MfTime,
	}
	if age.GetRealname() == membermdl.RealnameVerified {
		item.IsRealname = true
		if !age.After14 {
			item.AgeBand = "14-"
		} else if !age.After18 {
			item.AgeBand = "14-18"
		} else {
			item.AgeBand = "18+"
		}
	}
	return item
}
