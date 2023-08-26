package service

import (
	"context"
	"strconv"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

func (s *Service) FavTabShow(ctx context.Context, req *v1.FavTabShowReq) (*v1.FavTabShowResp, error) {
	if req.Mid <= 0 {
		return nil, ecode.RequestErr
	}
	ret := &v1.FavTabShowResp{}
	// 因为该接口涉及的流量入口非常明显
	// 如果发生错误就默认不展示该tab入口
	resp, err := s.dao.PersonalMenuStatusV1(ctx, dao.PersonalMenuStatusOpt{Mid: req.Mid})
	if err != nil {
		log.Errorc(ctx, "failed to query favTabShow. Discarded: %v", err)
		return ret, nil
	}
	dev, _, _ := DevNetAuthFromCtx(ctx)

	if s.C.Feature.MusicFavTabShow.Enabled(dev) {
		ret.ShowMenu = resp.HasMenu || resp.HasCollection || resp.HasMenuCreated
	} else {
		ret.ShowMenu = resp.HasMenu
	}
	return ret, nil
}

func (s *Service) MainFavMusicSubTabList(ctx context.Context, _ *v1.MainFavMusicSubTabListReq) (ret *v1.MainFavMusicSubTabListResp, err error) {
	_, _, auth := DevNetAuthFromCtx(ctx)

	st, err := s.dao.PersonalMenuStatusV1(ctx, dao.PersonalMenuStatusOpt{Mid: auth.Mid})
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to check user(%d) fav music subTab status: %v", auth.Mid, err)
	}

	type tabStat struct {
		Show bool
		Tab  *v1.MusicSubTab
	}
	// tabs是否显示/资源计数
	tabShow := map[int32]*tabStat{
		model.MenuCreated:       {},
		model.MenuFavored:       {},
		model.CollectionFavored: {},
	}
	for _, o := range []struct {
		Key  int32
		Show bool
	}{
		{model.MenuCreated, st.HasMenuCreated},
		{model.MenuFavored, st.HasMenu},
		{model.CollectionFavored, st.HasCollection},
	} {
		if o.Show {
			tabShow[o.Key].Show = true
		}
	}
	ret = new(v1.MainFavMusicSubTabListResp)
	var defaultTab int32
	// 创建的歌单>收藏的歌单>收藏的合辑
	tabPriority := []int32{model.MenuCreated, model.MenuFavored, model.CollectionFavored}
	for _, typ := range tabPriority {
		// 没内容不下发
		if s := tabShow[typ]; s == nil || !s.Show {
			continue
		}
		// 把第一个tab标记为默认tab
		if len(ret.Tabs) == 0 {
			defaultTab = typ
		}
		tab := &v1.MusicSubTab{
			TabType: typ, Name: model.MenuType2Name[typ],
		}
		tabShow[typ].Tab = tab
		ret.Tabs = append(ret.Tabs, tab)
	}
	if len(ret.Tabs) > 0 {
		ret.FirstPageRes = make(map[int32]*v1.MainFavMusicMenuListResp)
		eg := errgroup.WithContext(ctx)
		var pageNum int64 = 1
		mu := sync.Mutex{}
		for _, t := range ret.Tabs {
			typ := t.TabType
			eg.Go(func(c context.Context) error {
				list, err := s.dao.MusicMenuListV1(ctx, dao.MusicMenuListOpt{Typ: typ, Mid: auth.Mid, PageNum: pageNum})
				if err != nil {
					return errors.WithMessagef(err, "failed to get the first page menu list for tab(%d). Discarded", typ)
				}
				res := list.ToV1MainFavMusicMenuListResp(auth.Mid)
				// 创建的歌单-默认歌单特殊处理
				if typ == model.MenuCreated {
					var ok bool
					res.MenuList, ok = removeEmptyDefaultMenu(res.MenuList)
					if ok {
						list.Total -= 1 // Tab计数减1
					}
				}
				tabShow[typ].Tab.Total = list.Total
				mu.Lock()
				ret.FirstPageRes[typ] = res
				mu.Unlock()
				if typ == defaultTab {
					ret.DefaultTabRes = res
				}
				return nil
			})
		}
		if sideErr := eg.Wait(); sideErr != nil {
			log.Errorc(ctx, "%v", sideErr)
		}
	}

	return
}

func removeEmptyDefaultMenu(list []*v1.MusicMenu) (ret []*v1.MusicMenu, ok bool) {
	ret = list
	for i, m := range list {
		// 默认歌单且数量为空
		if m.IsDefaultMenu() && m.Total == 0 {
			listLen := len(list)
			list[i], list[listLen-1] = list[listLen-1], list[i]
			ret = list[0 : listLen-1]
			ok = true
		}
	}
	return
}

func (s *Service) MainFavMusicMenuList(ctx context.Context, req *v1.MainFavMusicMenuListReq) (ret *v1.MainFavMusicMenuListResp, err error) {
	if req.TabType <= 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "empty req.TabType")
	}
	if len(req.Offset) == 0 {
		// 默认从第一页开始
		req.Offset = "1"
	}
	pageNum, err := strconv.ParseInt(req.Offset, 10, 64)
	if err != nil {
		return nil, errors.WithMessagef(ecode.RequestErr, "unable to parse page offset")
	}
	_, _, auth := DevNetAuthFromCtx(ctx)

	list, err := s.dao.MusicMenuListV1(ctx, dao.MusicMenuListOpt{Typ: req.TabType, Mid: auth.Mid, PageNum: pageNum})
	ret = list.ToV1MainFavMusicMenuListResp(auth.Mid)
	ret.MenuList, _ = removeEmptyDefaultMenu(ret.MenuList)
	return
}

func (s *Service) MenuEdit(ctx context.Context, req *v1.MenuEditReq) (ret *v1.MenuEditResp, err error) {
	if req.Id <= 0 || len(req.Title) <= 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "empty req.Id or req.Title")
	}
	_, _, auth := DevNetAuthFromCtx(ctx)
	ret = new(v1.MenuEditResp)
	err = s.dao.MenuEditV1(ctx, dao.MenuEditOpt{
		Mid: auth.Mid, MenuId: req.Id, Title: req.Title, Desc: req.Desc, IsOpen: req.IsPublic,
	})
	if err == nil {
		ret.Message = s.C.Res.Text.EditOK
	}
	return
}

func (s *Service) MenuDelete(ctx context.Context, req *v1.MenuDeleteReq) (ret *v1.MenuDeleteResp, err error) {
	if req.Id <= 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "zero req.Id")
	}
	_, _, auth := DevNetAuthFromCtx(ctx)
	ret = new(v1.MenuDeleteResp)
	err = s.dao.MenuDelV1(ctx, dao.MenuDelOpt{MenuId: req.Id, Mid: auth.Mid})
	if err == nil {
		ret.Message = s.C.Res.Text.DeleteFavFolder
	}
	return
}

func (s *Service) MenuSubscribe(ctx context.Context, req *v1.MenuSubscribeReq) (ret *v1.MenuSubscribeResp, err error) {
	if req.TargetId <= 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "zero req.TargetId")
	}
	ret = new(v1.MenuSubscribeResp)
	_, _, auth := DevNetAuthFromCtx(ctx)

	switch req.Action {
	case v1.MenuSubscribeReq_ADD:
		err = s.dao.MenuCollAdd(ctx, dao.MenuCollAddOpt{MenuId: req.TargetId, Mid: auth.Mid})
		ret.Message = s.C.Res.Text.AddFav
	case v1.MenuSubscribeReq_DEL:
		err = s.dao.MenuCollDelV1(ctx, dao.MenuCollDelOpt{MenuId: req.TargetId, Mid: auth.Mid})
		ret.Message = s.C.Res.Text.DelFav
	default:
		return nil, errors.WithMessagef(ecode.RequestErr, "unknown req.Action(%v)", req.Action)
	}

	return
}

func (s *Service) Click(ctx context.Context, req *v1.ClickReq) (ret *v1.ClickResp, err error) {
	if req.Sid <= 0 || req.Action == v1.ClickReq_INVALID {
		return nil, errors.WithMessagef(ecode.RequestErr, "zero songId or invalid action")
	}
	ret = new(v1.ClickResp)
	err = s.dao.MusicClickReport(ctx, dao.MusicClickReportOpt{
		SongId: req.Sid, ClickTyp: dao.MusicClickShare, AddMetric: true,
	})
	return
}
