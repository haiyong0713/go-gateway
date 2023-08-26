package feature

import (
	"context"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	"go-gateway/app/app-svr/app-feed/admin/model/tree"

	"go-gateway/app/app-svr/app-feed/admin/model/feature"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func (s *Service) AppList(c context.Context, cookie string) (*feature.AppListRly, error) {
	var (
		treeNodes                    []*tree.Node
		err                          error
		buildLimitCount, abTestCount map[int]int
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() error {
		treeNodes, err = s.dao.FetchRoleTree(ctx, cookie)
		if err != nil {
			log.Error("s.dao.FetchRoleTree(%+v) (%+v)", cookie, err)
		}
		return err
	})
	g.Go(func() error {
		buildLimitCount, err = s.dao.BuildLimitServiceCount(ctx)
		if err != nil {
			log.Error("s.dao.BuildLimitServiceCount(%+v) (%+v)", cookie, err)
		}
		return nil
	})
	g.Go(func() error {
		abTestCount, err = s.dao.ABTestServiceCount(ctx)
		if err != nil {
			log.Error("s.dao.ABTestServiceCount(%+v) (%+v)", cookie, err)
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	// 组装数据
	var list, listEmpty []*feature.App
	// 抽象出一个全局服务
	commonApp := &feature.App{
		App:       "全局模式",
		TreeID:    feature.Common,
		Dimension: "common",
	}
	if count, ok := buildLimitCount[feature.Common]; ok {
		commonApp.Count = count
		delete(buildLimitCount, feature.Common)
	}
	if count, ok := abTestCount[feature.Common]; ok {
		commonApp.Count += count
		delete(abTestCount, feature.Common)
	}
	for _, node := range treeNodes {
		if node == nil {
			continue
		}
		app := &feature.App{}
		app.FormTreeNode(node)
		var count = buildLimitCount[app.TreeID]
		count += abTestCount[app.TreeID]
		if count > 0 {
			app.Count = count
			list = append(list, app)
			continue
		}
		listEmpty = append(listEmpty, app)
	}
	list = append([]*feature.App{commonApp}, append(list, listEmpty...)...)
	return &feature.AppListRly{
		Total: len(list),
		List:  list,
	}, nil
}

func (s *Service) AppPlat(ctx context.Context, req *feature.AppPlatReq) (*feature.AppPlatRly, error) {
	svrAttr, err := s.dao.GetSvrAttrByTreeID(ctx, req.TreeID)
	if err != nil && err != ecode.NothingFound {
		log.Error("s.dao.GetSvrAttrByTreeID(%+v) error(%+v)", req.TreeID, err)
		return nil, err
	}
	appPlats := s.dao.GroupAppPlats()
	if svrAttr != nil {
		mobiApps := util.PartMap(svrAttr.MobiApps, ",")
		for _, items := range appPlats {
			for _, item := range items {
				if _, ok := mobiApps[item.MobiApp]; !ok {
					continue
				}
				item.IsChosen = true
			}
		}
	}
	return &feature.AppPlatRly{List: appPlats}, nil
}

func (s *Service) SaveApp(ctx context.Context, userID int, username string, req *feature.SaveAppReq) error {
	if err := s.validateMobiApps(ctx, req.MobiApps); err != nil {
		log.Error("s.validateMobiApps(%+v) error(%+v)", req.MobiApps, err)
		return err
	}
	svrAttr, err := s.dao.GetSvrAttrByTreeID(ctx, req.TreeID)
	if err != nil && err != ecode.NothingFound {
		log.Error("s.dao.GetSvrAttrByTreeID(%+v) error(%+v)", req.TreeID, err)
		return err
	}
	attrs := &feature.ServiceAttribute{
		TreeID:      req.TreeID,
		MobiApps:    req.MobiApps,
		Modifier:    username,
		ModifierUID: userID,
	}
	if svrAttr != nil {
		attrs.ID = svrAttr.ID
		attrs.Ctime = svrAttr.Ctime
	}
	if err := s.dao.SaveSvrAttr(ctx, attrs); err != nil {
		log.Error("s.dao.SaveSvrAttr(%+v) error(%+v)", attrs, err)
		return err
	}
	return nil
}

func (s *Service) validateMobiApps(_ context.Context, mobiApps string) error {
	plats := s.dao.Plats()
	mobiAppList := strings.Split(mobiApps, ",")
	for _, mobiApp := range mobiAppList {
		if _, ok := plats[mobiApp]; !ok {
			return ecode.RequestErr
		}
	}
	return nil
}
