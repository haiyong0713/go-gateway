package hidden

import (
	"context"
	"encoding/json"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	locdao "go-gateway/app/app-svr/app-feed/admin/dao/location"
	"go-gateway/app/app-svr/app-feed/admin/dao/show"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/hidden"
	"go-gateway/app/app-svr/app-feed/admin/model/menu"
	showmdl "go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// Service is menu service
type Service struct {
	showDao *show.Dao
	locDao  *locdao.Dao
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		showDao: show.New(c),
		locDao:  locdao.New(c),
	}
	return
}

// HiddenList is
func (s *Service) HiddenList(c context.Context, arg *hidden.ListParam) (*hidden.ListReply, error) {
	hiddens, total, err := s.showDao.Hiddens(c, arg.Pn, arg.Ps)
	if err != nil || hiddens == nil {
		log.Error("s.showDao.Hiddens arg(%+v) err(%+v) or nil", arg, err)
		return nil, nil
	}
	var (
		oids, pids   []int64
		hiddenLimits map[int64][]*hidden.HiddenLimit
		areaMap      = make(map[int64][]int64)
	)
	for _, v := range hiddens {
		oids = append(oids, v.ID)
		pids = append(pids, v.PID)
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		hiddenLimits, err = s.showDao.HiddenLimits(c, oids)
		return err
	})
	eg.Go(func(ctx context.Context) (e error) {
		//nolint:ineffassign,staticcheck
		areaMap, e = s.locDao.PolicyInfos(ctx, pids)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Error("HiddenList eg.Wait() err(%+v)", err)
		return nil, err
	}
	pager := &hidden.ListReply{Page: &hidden.Page{Num: arg.Pn, Size: arg.Ps, Total: total}}
	//拼接信息
	for _, v := range hiddens {
		tmp := &hidden.HiddenInfo{Hidden: v}
		if hl, ok := hiddenLimits[v.ID]; ok {
			tmp.Limit = hl
		}
		if area, ok := areaMap[v.PID]; ok {
			tmp.AreaIDs = area
		}
		pager.List = append(pager.List, tmp)
	}
	return pager, nil
}

// HiddenDetail is
func (s *Service) HiddenDetail(c context.Context, id int64) (*hidden.HiddenInfo, error) {
	hd, err := s.showDao.Hidden(c, id)
	if err != nil || hd == nil {
		log.Error("s.showDao.Hidden id(%d) err(%+v) or hd=nil", id, err)
		return nil, err
	}
	var (
		tabLimits     map[int64][]*hidden.HiddenLimit
		sideBars      map[int64]*menu.Sidebar
		region        *hidden.Region
		moduleInfo    *showmdl.SidebarModule
		areaIDs, sids []int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		tabLimits, e = s.showDao.HiddenLimits(ctx, []int64{hd.ID})
		return
	})
	if hd.SID > 0 {
		sids = append(sids, hd.SID)
	}
	if hd.CID > 0 {
		sids = append(sids, hd.CID)
	}
	if len(sids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			sideBars, e = s.showDao.SideBars(ctx, sids)
			return e
		})
	}
	if hd.RID > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			region, e = s.showDao.Region(ctx, hd.RID)
			return e
		})
	}
	eg.Go(func(ctx context.Context) (e error) {
		areaIDs, e = s.locDao.PolicyInfo(ctx, hd.PID)
		return e
	})
	if hd.ModuleID > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			moduleInfo, e = s.showDao.Module(ctx, hd.ModuleID)
			return e
		})

	}
	if err = eg.Wait(); err != nil {
		log.Error("HiddenDetail eg.Wait err(%+v)", err)
		return nil, err
	}
	res := &hidden.HiddenInfo{Hidden: hd}
	res.HideDynamic = hd.HideDynamic
	if side, ok := sideBars[hd.SID]; ok {
		res.SName = side.Name
	}
	if ce, ok := sideBars[hd.CID]; ok {
		res.CName = ce.Name
	}
	if region != nil {
		res.RName = region.Name
	}
	if len(areaIDs) > 0 {
		res.AreaIDs = areaIDs
	}
	if moduleInfo != nil {
		res.MName = moduleInfo.Name
	}
	if lv, ok := tabLimits[hd.ID]; ok {
		res.Limit = lv
	}
	return res, nil
}

// SaveLog .
func (s *Service) SaveLog(c context.Context, uid, id int64, opAction, operator string) error {
	hiddenInfo, err := s.showDao.Hidden(c, id)
	if err != nil {
		log.Error("SaveLog s.showDao.Hidden (%d) error(%v) or nil", id, err)
		return err
	}
	hiddenLimits, err := s.showDao.HiddenLimits(c, []int64{id})
	if err != nil {
		log.Error("SaveLog s.showDao.HiddenLimits (%d) error(%v)", id, err)
		return err
	}
	arg := make(map[string]interface{})
	arg["info"] = hiddenInfo
	if lVal, ok := hiddenLimits[id]; ok {
		arg["limit"] = lVal
	}
	if err = util.AddLog(common.LogEntranceHidden, operator, uid, id, opAction, arg); err != nil {
		log.Error("SaveLog AddLog error(%v)", err)
		return err
	}
	return nil
}

// HiddenOpt is
func (s *Service) HiddenOpt(c context.Context, id, uid int64, state int, operator string) error {
	//下线操作
	whState := hidden.StateOnline
	upState := hidden.StateOffline
	opAction := common.ActionOffline
	if state == hidden.StateDel { //删除操作,下线状态才可以删除
		whState = hidden.StateOffline
		upState = hidden.StateDel
		opAction = common.ActionDelete
	} else if state == hidden.StateOnline { //上线
		whState = hidden.StateOffline
		upState = hidden.StateOnline
		opAction = common.ActionOnline
	}
	rows, err := s.showDao.UpdateHiddenState(c, id, whState, upState)
	if err != nil {
		log.Error("HiddenOpt s.showDao.UpdateHiddenState id:%d whState:%d upState:%d) error(%v)", id, whState, upState, err)
		return err
	}
	if rows == 0 {
		return ecode.NotModified
	}
	// 操作日志
	if err = s.SaveLog(c, uid, int64(id), opAction, operator); err != nil {
		log.Error("HiddenOpt s.SaveLog error(%v)", err)
		return err
	}
	return nil
}

// HiddenSave .
func (s *Service) HiddenSave(c context.Context, arg *hidden.HiddenSaveParam, operator string, uid int64) error {
	buildLimit := make([]*hidden.HiddenLimit, 0)
	limits := make(map[int8]*hidden.HiddenLimit)
	if err := json.Unmarshal([]byte(arg.Limit), &buildLimit); err != nil {
		log.Error("HiddenSave BuildLimit json.Unmarshal %s error(%v)", arg.Limit, err)
		return err
	}
	for _, v := range buildLimit {
		if err := v.ValidateParam(); err != nil {
			return err
		}
		limits[v.Plat] = v
	}
	pid, err := s.locDao.AddPolicy(c, arg.AreaIDs)
	if err != nil {
		log.Error("HiddenSave s.locDao.AddPolicy arg(%v) err(%+v) or pid=0", arg, err)
		return err
	}
	if pid <= 0 {
		log.Error("HiddenSave s.locDao.AddPolicy arg(%v) pid=0", arg)
		return ecode.RequestErr
	}
	hiddenParam := &hidden.Hidden{ID: arg.ID, SID: arg.SID, RID: arg.RID, CID: arg.CID, Channel: arg.Channel, PID: pid, Stime: arg.Stime, Etime: arg.Etime, HiddenCondition: arg.HiddenCondition, ModuleID: arg.ModuleID, HideDynamic: arg.HideDynamic}
	//nolint:ineffassign,staticcheck
	optID, err := s.showDao.HiddenSave(c, hiddenParam, limits)
	// 操作日志
	opAction := common.ActionAdd
	if arg.ID > 0 {
		opAction = common.ActionUpdate
	}
	if err = s.SaveLog(c, uid, optID, opAction, operator); err != nil {
		log.Error("HiddenSave s.SaveLog error(%v)", err)
		return err
	}
	return nil
}

// EntranceSearch .
func (s *Service) EntranceSearch(c context.Context, oid int64, oType int) (*hidden.Entrance, error) {
	switch oType {
	case hidden.EntranceHome, hidden.EntranceSidebar:
		sideBars, err := s.showDao.SideBars(c, []int64{oid})
		if err != nil {
			log.Error("EntranceSearch s.showDao.SideBars oid(%d) err(%+v)", oid, err)
			return nil, err
		}
		sv, ok := sideBars[oid]
		if !ok || sv == nil || (int8(sv.Plat) != showmdl.PlatAndroid && int8(sv.Plat) != showmdl.PlatAndroidI) {
			return nil, ecode.NothingFound
		}
		return &hidden.Entrance{ID: sv.ID, Title: sv.Name, Plat: sv.Plat}, nil
	case hidden.EntranceChannel:
		region, err := s.showDao.Region(c, oid)
		if err != nil {
			log.Error("EntranceSearch s.showDao.Region oid(%d) err(%v) or nil", oid, err)
			return nil, err
		}
		if region.ID <= 0 { // 所有平台分区rid都一样
			return nil, ecode.NothingFound
		}
		return &hidden.Entrance{ID: region.ID, Title: region.Name, Plat: region.Plat}, err
	case hidden.EntranceModule:
		moduleInfo, err := s.showDao.Module(c, oid)
		if err != nil {
			log.Error("EntranceSearch s.showDao.Module oid(%d) err(%+v)", oid, err)
			return nil, err
		}
		return &hidden.Entrance{ID: moduleInfo.ID, Title: moduleInfo.Title, Plat: int(moduleInfo.Plat)}, nil
	default:
		return nil, ecode.NothingFound
	}
}
