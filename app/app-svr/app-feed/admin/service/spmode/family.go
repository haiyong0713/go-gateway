package spmode

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	familymdl "go-gateway/app/app-svr/app-feed/admin/model/family"
	spmodemdl "go-gateway/app/app-svr/app-feed/admin/model/spmode"
)

func (s *Service) SearchFamily(ctx context.Context, req *familymdl.SearchFamilyReq) (*familymdl.SearchFamilyRly, error) {
	eg := errgroup.WithContext(ctx)
	var parentRels []*familymdl.FamilyRelation
	eg.Go(func(ctx context.Context) error {
		if rly, err := s.dao.FamilyRelsOfParent(req.Mid); err == nil {
			parentRels = rly
		}
		return nil
	})
	var childRel *familymdl.FamilyRelation
	eg.Go(func(ctx context.Context) error {
		if rly, err := s.dao.FamilyRelOfChild(req.Mid); err == nil {
			childRel = rly
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Fail to fetch SearchFamily material, mid=%+v error=%+v", req.Mid, err)
		return nil, ecode.Error(ecode.ServerErr, "数据查询失败")
	}
	var (
		err  error
		item *familymdl.SearchFamilyItem
	)
	switch identity(parentRels, childRel) {
	case familymdl.IdentityParent:
		item, err = s.searchParent(ctx, parentRels)
	case familymdl.IdentityChild:
		item, err = s.searchChild(ctx, childRel)
	case familymdl.IdentityNormal:
		item = s.searchNormal(ctx, req.Mid)
	default:
		return nil, ecode.Error(ecode.ServerErr, "亲子关系异常")
	}
	if err != nil {
		return nil, err
	}
	return &familymdl.SearchFamilyRly{List: []*familymdl.SearchFamilyItem{item}}, nil
}

func (s *Service) searchParent(ctx context.Context, rels []*familymdl.FamilyRelation) (*familymdl.SearchFamilyItem, error) {
	if len(rels) == 0 {
		return nil, ecode.NothingFound
	}
	pmid := rels[0].ParentMid
	var mids = []int64{pmid}
	for _, rel := range rels {
		if rel == nil || rel.ChildMid <= 0 {
			continue
		}
		mids = append(mids, rel.ChildMid)
	}
	var (
		users      []*familymdl.RelatedUser
		parentName string
	)
	if accounts, err := s.accountDao.Infos3(ctx, mids); err == nil {
		parentName = accounts[pmid].GetName()
		for _, rel := range rels {
			if rel == nil {
				continue
			}
			user := &familymdl.RelatedUser{ID: rel.ID, Mid: rel.ChildMid}
			if value, ok := accounts[rel.ChildMid]; ok && value != nil {
				user.UserName = value.GetName()
			}
			users = append(users, user)
		}
	}
	return &familymdl.SearchFamilyItem{
		Identity:     familymdl.IdentityParent,
		Mid:          pmid,
		UserName:     parentName,
		RelatedUsers: users,
	}, nil
}

func (s *Service) searchChild(ctx context.Context, rel *familymdl.FamilyRelation) (*familymdl.SearchFamilyItem, error) {
	if rel == nil {
		return nil, ecode.NothingFound
	}
	mids := []int64{rel.ChildMid, rel.ParentMid}
	accounts, _ := s.accountDao.Infos3(ctx, mids)
	return &familymdl.SearchFamilyItem{
		Identity: familymdl.IdentityChild,
		Mid:      rel.ChildMid,
		UserName: accounts[rel.ChildMid].GetName(),
		RelatedUsers: []*familymdl.RelatedUser{
			{
				ID:       rel.ID,
				Mid:      rel.ParentMid,
				UserName: accounts[rel.ParentMid].GetName(),
			},
		},
	}, nil
}

func (s *Service) searchNormal(ctx context.Context, mid int64) *familymdl.SearchFamilyItem {
	accounts, _ := s.accountDao.Infos3(ctx, []int64{mid})
	return &familymdl.SearchFamilyItem{
		Identity: familymdl.IdentityNormal,
		Mid:      mid,
		UserName: accounts[mid].GetName(),
	}
}

func (s *Service) BindList(ctx context.Context, req *familymdl.BindListReq) (*familymdl.BindListRly, error) {
	total, list, err := s.dao.PagingFamilyLog(req.Mid, req.Pn, req.Ps)
	if err != nil {
		return nil, ecode.Error(ecode.ServerErr, "数据查询失败")
	}
	return &familymdl.BindListRly{
		Page: &familymdl.Page{Num: req.Pn, Size: req.Ps, Total: total},
		List: list,
	}, nil
}

func (s *Service) UnbindFamily(ctx context.Context, req *familymdl.UnbindReq, userid int64, username string) error {
	rel, err := s.dao.FamilyRelById(req.ID)
	if err != nil {
		return ecode.Error(ecode.ServerErr, "亲子关系查询失败")
	}
	if rel == nil || rel.State == familymdl.StateUnbind {
		return nil
	}
	teen, err := s.dao.TeenagerUserByMidModel(rel.ChildMid, spmodemdl.ModelTeenager)
	if err != nil {
		return ecode.Error(ecode.ServerErr, "青少年状态查询失败")
	}
	// 解绑
	if err := s.dao.UnbindFamily(rel.ID); err != nil {
		return ecode.Error(ecode.ServerErr, "解绑失败")
	}
	_ = s.worker.Do(ctx, func(ctx context.Context) {
		_ = s.dao.DelCacheFamilyRelsOfParent(ctx, rel.ParentMid)
		_ = s.dao.DelCacheFamilyRelsOfChild(ctx, rel.ChildMid)
		_ = s.dao.BatchAddFamilyLogs([]*familymdl.FamilyLog{
			{Mid: rel.ParentMid, Operator: username, Content: fmt.Sprintf("系统解绑与%d的亲子关系", rel.ChildMid)},
			{Mid: rel.ChildMid, Operator: username, Content: fmt.Sprintf("系统解绑与%d的亲子关系", rel.ParentMid)},
		})
	})
	// 解除青少年模式
	if teen != nil && teen.State == spmodemdl.StateOpen {
		return s.relieveUserFromFamily(ctx, teen.ID, userid, username)
	}
	return nil
}

func (s *Service) relieveUserFromFamily(ctx context.Context, id, userid int64, username string) error {
	ok, err := s.relieveUser(ctx, id, spmodemdl.OperationQuitFyMgrUnbind)
	if err != nil {
		return ecode.Error(ecode.ServerErr, "青少年模式解除失败")
	}
	if ok {
		_ = s.worker.Do(ctx, func(ctx context.Context) {
			_ = s.dao.AddSpecialModeLog(&spmodemdl.SpecialModeLog{
				RelatedKey:  fmt.Sprintf("%s_%d", spmodemdl.RelatedKeyTypeUser, id),
				OperatorUid: userid,
				Operator:    username,
				Content:     "亲子平台后台解绑退出",
			})
		})
	}
	return nil
}

func identity(parentRels []*familymdl.FamilyRelation, childRel *familymdl.FamilyRelation) string {
	if len(parentRels) > 0 && childRel == nil {
		return familymdl.IdentityParent
	}
	if len(parentRels) == 0 && childRel != nil {
		return familymdl.IdentityChild
	}
	if len(parentRels) > 0 && childRel != nil {
		return familymdl.IdentityAbnormal
	}
	return familymdl.IdentityNormal
}
